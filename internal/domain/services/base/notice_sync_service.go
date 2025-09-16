package base_services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// processResult represents the result of processing a single notice
type processResult struct {
	noticeID     int
	md5          string
	success      bool
	syncRequired bool
	error        error
}

const (
	minWorkers       = 10 // increase min worker count
	maxWorkers       = 30 // increase max worker count
	noticesPerWorker = 3  // decrease notices per worker for better concurrency

	// Building concurrency control
	minBuildingWorkers       = 1    // min building worker
	maxBuildingWorkerPerCore = 20   // max building worker per core
	buildingWorkerLoadFactor = 0.75 // building worker load factor (0-1)

	// Redis cache keys
	syncedNoticeIDsPrefix = "building_synced_notice_ids" // 用于存储上次同步的旧系统通知ID和MD5
	syncedNoticeMD5Prefix = "building_synced_notice_md5" // 用于存储上次同步的通知MD5列表
)

// 定义同步通知信息结构体
type SyncedNoticeInfo struct {
	ID  int    `json:"id"`
	MD5 string `json:"md5"`
}

func getNoticeSyncInterval() time.Duration {
	interval := os.Getenv("NOTICE_SYNC_INTERVAL")
	if interval == "" {
		return 2 * time.Minute // default to 2 minutes if not set
	}

	intervalInt, err := strconv.Atoi(interval)
	if err != nil {
		return 2 * time.Minute // default to 2 minutes if invalid value
	}

	return time.Duration(intervalInt) * time.Minute
}

// notice list redis cache duration
func getNoticeSyncCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_CACHE_DURATION")
	if duration == "" {
		return 4 * time.Minute // default to 4 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 4 * time.Minute // default to 4 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

// building redis
func getNoticeSyncBuildingCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_BUILDING_CACHE_DURATION")
	if duration == "" {
		return 30 * time.Minute // default to 30 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 30 * time.Minute // default to 30 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

func getNoticeSyncCountCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_COUNT_CACHE_DURATION")
	if duration == "" {
		return 60 * time.Minute // default to 60 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 60 * time.Minute // default to 60 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

// calculateWorkerCount dynamically calculate worker count based on notice count
func calculateWorkerCount(noticeCount int) int {
	// Calculate suggested worker count based on notice count and available CPU cores
	cpuCount := runtime.NumCPU()
	suggestedWorkers := noticeCount / noticesPerWorker

	// Scale based on CPU cores
	maxWorkersForCPU := cpuCount * 4

	if suggestedWorkers < minWorkers {
		return minWorkers
	}
	if suggestedWorkers > maxWorkers || suggestedWorkers > maxWorkersForCPU {
		if maxWorkers < maxWorkersForCPU {
			return maxWorkers
		}
		return maxWorkersForCPU
	}
	return suggestedWorkers
}

// Add new function for calculating building workers
func calculateBuildingWorkers(buildingCount int, cpuCores int) int {
	// Calculate max concurrency based on CPU cores
	maxWorkers := int(float64(cpuCores) * maxBuildingWorkerPerCore * buildingWorkerLoadFactor)

	// Ensure at least minBuildingWorkers
	if maxWorkers < minBuildingWorkers {
		maxWorkers = minBuildingWorkers
	}

	// If building count is less than max workers, use building count
	if buildingCount < maxWorkers {
		return buildingCount
	}

	return maxWorkers
}

type InterfaceNoticeSyncService interface {
	SyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error)
	StartSyncScheduler(ctx context.Context)
	ManualSyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error)
}

type NoticeSyncService struct {
	db              *gorm.DB
	redis           *redis.Client
	buildingService InterfaceBuildingService
	uploadService   IUploadService
	fileService     InterfaceFileService
}

func NewNoticeSyncService(db *gorm.DB, redis *redis.Client, buildingService InterfaceBuildingService, uploadService IUploadService, fileService InterfaceFileService) InterfaceNoticeSyncService {
	return &NoticeSyncService{
		db:              db,
		redis:           redis,
		buildingService: buildingService,
		uploadService:   uploadService,
		fileService:     fileService,
	}
}

type OldSystemNotice struct {
	ID        int    `json:"id"`
	MessTitle string `json:"mess_title"`
	MessType  string `json:"mess_type"`
	MessFile  string `json:"mess_file"`
}

// 1. getCachedNoticeIDs
func (s *NoticeSyncService) getCachedNoticeIDs(ctx context.Context, buildingID uint) ([]int, error) {
	key := fmt.Sprintf("building_notice_ids:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// 2. setCachedNoticeIDs
func (s *NoticeSyncService) setCachedNoticeIDs(ctx context.Context, buildingID uint, ids []int) error {
	key := fmt.Sprintf("building_notice_ids:%d", buildingID)
	jsonIDs, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, string(jsonIDs), getNoticeSyncCacheDuration()).Err()
}

// 3. getCachedBuilding
func (s *NoticeSyncService) getCachedBuilding(ctx context.Context, buildingID uint) (*base_models.Building, error) {
	key := fmt.Sprintf("building:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss, get from DB and cache it
		var building base_models.Building
		if err := s.db.First(&building, buildingID).Error; err != nil {
			return nil, err
		}

		// Cache building data
		jsonData, err := json.Marshal(building)
		if err != nil {
			return nil, err
		}
		if err := s.redis.Set(ctx, key, string(jsonData), getNoticeSyncBuildingCacheDuration()).Err(); err != nil {
			return nil, err
		}
		return &building, nil
	} else if err != nil {
		return nil, err
	}

	var building base_models.Building
	if err := json.Unmarshal([]byte(val), &building); err != nil {
		return nil, err
	}
	return &building, nil
}

// 4. getCachedNoticeCount
func (s *NoticeSyncService) getCachedNoticeCount(ctx context.Context, buildingID uint) (int64, error) {
	key := fmt.Sprintf("building_ismart_notice_count:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss, get from DB and cache it
		var totalIsmartNotices int64
		if err := s.db.Model(&base_models.Notice{}).
			Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
			Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", buildingID, true).
			Count(&totalIsmartNotices).Error; err != nil {
			return 0, fmt.Errorf("failed to get total ismart notices: %v", err)
		}

		// Cache the count for 1 hour
		if err := s.redis.Set(ctx, key, totalIsmartNotices, time.Hour).Err(); err != nil {
			return 0, err
		}
		return totalIsmartNotices, nil
	} else if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 5. updateCachedNoticeCount
func (s *NoticeSyncService) updateCachedNoticeCount(ctx context.Context, buildingID uint, count int64) error {
	key := fmt.Sprintf("building_ismart_notice_count:%d", buildingID)
	return s.redis.Set(ctx, key, count, getNoticeSyncCountCacheDuration()).Err()
}

// 修改获取上次同步的旧系统通知信息函数
func (s *NoticeSyncService) getSyncedNoticeInfos(ctx context.Context, buildingID uint) ([]SyncedNoticeInfo, error) {
	key := fmt.Sprintf("%s:%d", syncedNoticeIDsPrefix, buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return []SyncedNoticeInfo{}, nil // 返回空数组而非nil，表示没有缓存数据
	} else if err != nil {
		return nil, err
	}

	var infos []SyncedNoticeInfo
	if err := json.Unmarshal([]byte(val), &infos); err != nil {
		return nil, err
	}
	return infos, nil
}

// 修改保存本次同步的旧系统通知信息函数
func (s *NoticeSyncService) setSyncedNoticeInfos(ctx context.Context, buildingID uint, infos []SyncedNoticeInfo) error {
	key := fmt.Sprintf("%s:%d", syncedNoticeIDsPrefix, buildingID)
	jsonInfos, err := json.Marshal(infos)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, string(jsonInfos), getNoticeSyncCacheDuration()).Err()
}

// 修改SyncBuildingNotices函数
func (s *NoticeSyncService) SyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error) {
	ctx := context.Background()

	// 获取建筑信息
	building, err := s.getCachedBuilding(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %v", err)
	}

	// 获取缓存的通知ID
	cachedIDs, err := s.getCachedNoticeIDs(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached notice IDs: %v", err)
	}
	log.Info("Redis缓存通知ID | 建筑ID: %d | 数量: %d", buildingID, len(cachedIDs))

	// 获取上次同步的通知MD5列表
	syncedMD5s, err := s.getSyncedNoticeMD5s(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get synced notice MD5s: %v", err)
	}
	log.Info("Redis缓存MD5数量 | 建筑ID: %d | 数量: %d", buildingID, len(syncedMD5s))

	// 获取上次同步的通知详细信息（包含ID和MD5）
	syncedNoticeInfos, err := s.getSyncedNoticeInfos(ctx, buildingID)
	if err != nil {
		log.Warn("获取上次同步通知详细信息失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		// 继续执行，不中断同步流程
	} else {
		log.Info("获取上次同步通知详细信息 | 建筑ID: %d | 数量: %d", buildingID, len(syncedNoticeInfos))
	}

	// 获取现有的iSmart通知
	var existingNotices []base_models.Notice
	if err := s.db.Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", buildingID, true).
		Preload("File").Find(&existingNotices).Error; err != nil {
		return nil, fmt.Errorf("failed to get existing notices: %v", err)
	}
	log.Info("数据库中找到现有通知 | 建筑ID: %d | 数量: %d", buildingID, len(existingNotices))

	// 创建现有通知的MD5映射
	existingMD5Map := make(map[string]base_models.Notice)
	for _, notice := range existingNotices {
		if notice.File != nil {
			existingMD5Map[notice.File.Md5] = notice
		}
	}

	// 请求旧系统的通知
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to request old system: %v", err)
	}
	defer resp.Body.Close()

	var oldNotices []OldSystemNotice
	if err := json.NewDecoder(resp.Body).Decode(&oldNotices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// 提取旧系统通知ID
	newIDs := make([]int, len(oldNotices))
	for i, notice := range oldNotices {
		newIDs[i] = notice.ID
	}
	log.Info("旧系统通知ID | 建筑ID: %d | 数量: %d", buildingID, len(newIDs))

	// 更新同步的通知详细信息
	var newSyncedInfos []SyncedNoticeInfo
	for _, notice := range oldNotices {
		// 这里假设我们会在后续处理中计算MD5
		// 实际实现中，应该在处理通知时计算MD5并更新此列表
		newSyncedInfos = append(newSyncedInfos, SyncedNoticeInfo{
			ID: notice.ID,
			// 暂时使用空MD5，实际实现中应该设置正确的值
			MD5: "",
		})
	}

	// 保存本次同步的通知详细信息
	if err := s.setSyncedNoticeInfos(ctx, buildingID, newSyncedInfos); err != nil {
		log.Warn("保存同步通知详细信息失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		// 继续执行，不中断同步流程
	} else {
		log.Info("保存同步通知详细信息 | 建筑ID: %d | 数量: %d", buildingID, len(newSyncedInfos))
	}

	// 检查是否需要强制同步
	totalIsmartNotices, err := s.getCachedNoticeCount(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached notice count: %v", err)
	}

	log.Info("数据库有%d个iSmart通知，旧系统有%d个通知", totalIsmartNotices, len(oldNotices))

	// 如果通知数量不匹配，强制同步
	if totalIsmartNotices != int64(len(oldNotices)) {
		log.Warn("通知数量不匹配，强制同步 | 建筑ID: %d", buildingID)
		// 清除缓存以强制同步
		if err := s.redis.Del(ctx, fmt.Sprintf("building_notice_ids:%d", buildingID)).Err(); err != nil {
			return nil, fmt.Errorf("failed to clear notice IDs cache: %v", err)
		}
		if err := s.redis.Del(ctx, fmt.Sprintf("building_ismart_notice_count:%d", buildingID)).Err(); err != nil {
			return nil, fmt.Errorf("failed to clear notice count cache: %v", err)
		}
		cachedIDs = nil
	}

	// 计算最佳工作线程数
	workers := calculateWorkerCount(len(oldNotices))
	resultChan := make(chan processResult, len(oldNotices))
	semaphore := make(chan struct{}, workers)

	log.Info("开始并发处理通知 | 工作线程: %d | 通知数: %d | 建筑ID: %d",
		workers, len(oldNotices), buildingID)

	// 并发处理通知，获取所有MD5值
	for _, oldNotice := range oldNotices {
		go func(notice OldSystemNotice) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			result := processResult{noticeID: notice.ID}

			// 下载并处理文件
			fileResp, err := http.Get(notice.MessFile)
			if err != nil {
				result.error = fmt.Errorf("failed to download file: %v", err)
				resultChan <- result
				return
			}
			defer fileResp.Body.Close()

			fileContent, err := io.ReadAll(fileResp.Body)
			if err != nil {
				result.error = fmt.Errorf("failed to read file content: %v", err)
				resultChan <- result
				return
			}

			md5Hash := md5.Sum(fileContent)
			md5Str := hex.EncodeToString(md5Hash[:])
			result.md5 = md5Str

			// 先不进行处理，只收集MD5
			result.success = true
			result.syncRequired = false
			resultChan <- result
		}(oldNotice)
	}

	// 收集所有新通知的MD5值
	var newMD5s []string
	var failedNotices []string
	md5ToNoticeID := make(map[string]int)

	for i := 0; i < len(oldNotices); i++ {
		result := <-resultChan
		if result.error != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", result.noticeID, result.error))
			continue
		}

		newMD5s = append(newMD5s, result.md5)
		md5ToNoticeID[result.md5] = result.noticeID
	}

	log.Info("从旧系统收集MD5值 | 数量: %d | 建筑ID: %d", len(newMD5s), buildingID)

	// 如果缓存的MD5列表存在且与新MD5列表完全匹配，跳过处理
	if len(syncedMD5s) > 0 && len(syncedMD5s) == len(newMD5s) {
		// 创建MD5集合用于快速查找
		syncedMD5Set := make(map[string]bool)
		for _, md5 := range syncedMD5s {
			syncedMD5Set[md5] = true
		}

		allMatch := true
		for _, md5 := range newMD5s {
			if !syncedMD5Set[md5] {
				allMatch = false
				break
			}
		}

		if allMatch {
			log.Info("MD5列表完全匹配，跳过同步 | 建筑ID: %d", buildingID)
			var hasSyncNotices []int
			for _, notice := range existingNotices {
				if notice.File != nil {
					hasSyncNotices = append(hasSyncNotices, int(notice.ID))
				}
			}

			// 更新缓存的通知ID
			if err := s.setCachedNoticeIDs(ctx, buildingID, newIDs); err != nil {
				return nil, fmt.Errorf("failed to update cached notice IDs: %v", err)
			}

			return gin.H{
				"message":           "No changes detected",
				"successCount":      0,
				"hasSyncedCount":    len(existingNotices),
				"deleteCount":       0,
				"failedNotices":     []string{},
				"totalProcessed":    len(oldNotices),
				"has_sync_notices":  hasSyncNotices,
				"need_sync_notices": []int{},
				"change_notices":    []int{},
			}, nil
		}
	}

	// 比较新旧MD5列表，找出需要添加和删除的通知

	// 1. 找出需要删除的通知（在旧MD5列表中存在但在新MD5列表中不存在）
	var md5sToDelete []string
	if len(syncedMD5s) > 0 {
		newMD5Set := make(map[string]bool)
		for _, md5 := range newMD5s {
			newMD5Set[md5] = true
		}

		for _, md5 := range syncedMD5s {
			if !newMD5Set[md5] {
				md5sToDelete = append(md5sToDelete, md5)
			}
		}
	}

	// 2. 找出需要添加的通知（在新MD5列表中存在但在旧MD5列表中不存在或在现有通知中不存在）
	var md5sToAdd []string
	syncedMD5Set := make(map[string]bool)
	for _, md5 := range syncedMD5s {
		syncedMD5Set[md5] = true
	}

	for _, md5 := range newMD5s {
		if !syncedMD5Set[md5] || existingMD5Map[md5].ID == 0 {
			md5sToAdd = append(md5sToAdd, md5)
		}
	}

	log.Info("发现需要删除的MD5: %d个, 需要添加的MD5: %d个 | 建筑ID: %d",
		len(md5sToDelete), len(md5sToAdd), buildingID)

	// 处理需要删除的通知
	deleteCount := 0
	if len(md5sToDelete) > 0 {
		for _, md5 := range md5sToDelete {
			if notice, exists := existingMD5Map[md5]; exists {
				// 解绑通知
				if err := s.unbindNotice(buildingID, notice.ID, &failedNotices); err != nil {
					log.Error("解绑通知失败 | MD5: %s | 错误: %v | 建筑ID: %d", md5, err, buildingID)
				} else {
					deleteCount++
				}
			}
		}
	}

	// 处理需要添加的通知
	var successCount, hasSyncedCount int
	var needSyncNotices, hasSyncNotices, changeNotices []int

	if len(md5sToAdd) > 0 {
		// 重新创建信号量和结果通道
		addWorkers := calculateWorkerCount(len(md5sToAdd))
		addResultChan := make(chan processResult, len(md5sToAdd))
		addSemaphore := make(chan struct{}, addWorkers)

		for _, md5 := range md5sToAdd {
			noticeID := md5ToNoticeID[md5]

			// 找到对应的旧系统通知
			var oldNotice OldSystemNotice
			for _, notice := range oldNotices {
				if notice.ID == noticeID {
					oldNotice = notice
					break
				}
			}

			if oldNotice.ID == 0 {
				failedNotices = append(failedNotices, fmt.Sprintf("Could not find old notice for MD5 %s", md5))
				continue
			}

			go func(notice OldSystemNotice, md5 string) {
				addSemaphore <- struct{}{}
				defer func() { <-addSemaphore }()

				result := processResult{noticeID: notice.ID, md5: md5}

				// 检查是否已存在相同MD5的通知
				if existingNotice, exists := existingMD5Map[md5]; exists {
					// 通知已存在，检查是否已绑定到当前建筑物
					var count int64
					if err := s.db.Table("notice_buildings").
						Where("notice_id = ? AND building_id = ?", existingNotice.ID, buildingID).
						Count(&count).Error; err != nil {
						result.error = fmt.Errorf("failed to check notice binding: %v", err)
					} else if count > 0 {
						// 通知已绑定，无需操作
						result.success = true
						result.syncRequired = false
						hasSyncNotices = append(hasSyncNotices, int(existingNotice.ID))
						hasSyncedCount++
					} else {
						// 通知存在但未绑定，需要绑定
						if err := s.db.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)",
							existingNotice.ID, buildingID).Error; err != nil {
							result.error = fmt.Errorf("failed to bind existing notice: %v", err)
						} else {
							result.success = true
							result.syncRequired = true
							needSyncNotices = append(needSyncNotices, int(existingNotice.ID))
							successCount++
						}
					}
				} else {
					// 通知不存在，需要下载并处理
					fileResp, err := http.Get(notice.MessFile)
					if err != nil {
						result.error = fmt.Errorf("failed to download file: %v", err)
						addResultChan <- result
						return
					}
					defer fileResp.Body.Close()

					fileContent, err := io.ReadAll(fileResp.Body)
					if err != nil {
						result.error = fmt.Errorf("failed to read file content: %v", err)
						addResultChan <- result
						return
					}

					// 处理通知
					if err := s.processNotice(buildingID, notice, fileContent, claims); err != nil {
						result.error = err
					} else {
						result.success = true
						result.syncRequired = true
						needSyncNotices = append(needSyncNotices, notice.ID)
						successCount++
					}
				}

				addResultChan <- result
			}(oldNotice, md5)
		}

		// 收集处理结果
		for i := 0; i < len(md5sToAdd); i++ {
			result := <-addResultChan
			if result.error != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", result.noticeID, result.error))
			}
		}
	}

	log.Info("同步完成 | 建筑ID: %d | 添加: %d | 删除: %d", buildingID, successCount, deleteCount)

	// 更新缓存
	if err := s.setCachedNoticeIDs(ctx, buildingID, newIDs); err != nil {
		return nil, fmt.Errorf("failed to update cached notice IDs: %v", err)
	}

	// 更新同步的通知MD5列表
	if err := s.setSyncedNoticeMD5s(ctx, buildingID, newMD5s); err != nil {
		return nil, fmt.Errorf("failed to update synced notice MD5s: %v", err)
	}

	// 更新缓存的通知数量
	newCount := int64(len(oldNotices))
	if err := s.updateCachedNoticeCount(ctx, buildingID, newCount); err != nil {
		log.Warn("警告: 更新缓存通知数量失败 | 建筑ID: %d | 错误: %v", buildingID, err)
	}

	// 同步设备轮播列表
	if err := s.syncAllDeviceNoticeCarousels(ctx, buildingID, needSyncNotices, changeNotices); err != nil {
		log.Warn("警告: 设备轮播同步失败 | 建筑ID: %d | 错误: %v", buildingID, err)
	} else {
		log.Info("设备轮播同步完成 | 建筑ID: %d | 新增: %d | 删除: %d",
			buildingID, len(needSyncNotices), len(changeNotices))
	}

	return gin.H{
		"message":           "Sync completed",
		"successCount":      successCount,
		"hasSyncedCount":    hasSyncedCount,
		"deleteCount":       deleteCount,
		"failedNotices":     failedNotices,
		"totalProcessed":    len(oldNotices),
		"has_sync_notices":  hasSyncNotices,
		"need_sync_notices": needSyncNotices,
		"change_notices":    changeNotices,
	}, nil
}

// 添加新函数：获取上次同步的通知MD5列表
func (s *NoticeSyncService) getSyncedNoticeMD5s(ctx context.Context, buildingID uint) ([]string, error) {
	key := fmt.Sprintf("%s:%d", syncedNoticeMD5Prefix, buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return []string{}, nil // 返回空数组而非nil，表示没有缓存数据
	} else if err != nil {
		return nil, err
	}

	var md5s []string
	if err := json.Unmarshal([]byte(val), &md5s); err != nil {
		return nil, err
	}
	return md5s, nil
}

// 添加新函数：保存本次同步的通知MD5列表
func (s *NoticeSyncService) setSyncedNoticeMD5s(ctx context.Context, buildingID uint, md5s []string) error {
	key := fmt.Sprintf("%s:%d", syncedNoticeMD5Prefix, buildingID)
	jsonMD5s, err := json.Marshal(md5s)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, string(jsonMD5s), getNoticeSyncCacheDuration()).Err()
}

// 添加新函数：解绑通知
func (s *NoticeSyncService) unbindNotice(buildingID uint, noticeID uint, failedNotices *[]string) error {
	// 开始事务
	tx := s.db.Begin()

	// 解绑通知与建筑物
	if err := tx.Exec("DELETE FROM notice_buildings WHERE notice_id = ? AND building_id = ?",
		noticeID, buildingID).Error; err != nil {
		tx.Rollback()
		*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to unbind notice ID %d: %v", noticeID, err))
		return err
	}

	log.Info("解绑通知ID | 通知ID: %d | 建筑ID: %d", noticeID, buildingID)

	// 检查通知是否绑定到其他建筑物
	var buildingCount int64
	if err := tx.Table("notice_buildings").Where("notice_id = ?", noticeID).Count(&buildingCount).Error; err != nil {
		tx.Rollback()
		*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to check notice bindings for ID %d: %v", noticeID, err))
		return err
	}

	if buildingCount == 0 {
		log.Info("通知ID %d 没有其他建筑绑定，将被删除", noticeID)

		// 获取通知信息，包括文件ID
		var notice base_models.Notice
		if err := tx.First(&notice, noticeID).Error; err != nil {
			tx.Rollback()
			*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to get notice ID %d: %v", noticeID, err))
			return err
		}

		// 保存文件ID以便后续检查
		fileID := notice.FileID

		// 删除通知
		if err := tx.Delete(&notice).Error; err != nil {
			tx.Rollback()
			*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to delete notice ID %d: %v", noticeID, err))
			return err
		}

		log.Info("删除通知ID | 通知ID: %d", noticeID)

		// 检查文件是否被其他通知使用
		if fileID != nil {
			var fileCount int64
			if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", *fileID).Count(&fileCount).Error; err != nil {
				tx.Rollback()
				*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to check file references for notice ID %d: %v", noticeID, err))
				return err
			}

			if fileCount == 0 {
				log.Info("文件ID %d 没有其他通知引用，将被删除", *fileID)

				// 删除文件
				if err := tx.Delete(&base_models.File{}, *fileID).Error; err != nil {
					tx.Rollback()
					*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to delete file for notice ID %d: %v", noticeID, err))
					return err
				}

				log.Info("删除文件ID | 文件ID: %d", *fileID)
			} else {
				log.Info("文件ID %d 有 %d 个其他通知引用，保留文件", *fileID, fileCount)
			}
		}
	} else {
		log.Info("通知ID %d 有 %d 个其他建筑绑定，保留通知", noticeID, buildingCount)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to commit unbind for notice ID %d: %v", noticeID, err))
		return err
	}

	return nil
}

// 添加设备轮播同步相关方法

// syncDeviceNoticeCarousel 同步设备通知轮播列表
func (s *NoticeSyncService) syncDeviceNoticeCarousel(ctx context.Context, buildingID uint, noticeMD5 string, noticeID uint, action string) error {
	// 获取绑定到该建筑物的所有设备
	var devices []base_models.Device
	if err := s.db.Where("building_id = ?", buildingID).Find(&devices).Error; err != nil {
		return fmt.Errorf("failed to get building devices: %v", err)
	}

	log.Info("开始同步设备轮播列表 | 建筑ID: %d | 通知MD5: %s | 操作: %s | 设备数量: %d",
		buildingID, noticeMD5, action, len(devices))

	for _, device := range devices {
		if err := s.updateDeviceNoticeCarousel(ctx, device.ID, noticeMD5, noticeID, action); err != nil {
			log.Warn("设备轮播同步失败 | 设备ID: %d | 建筑ID: %d | 错误: %v",
				device.ID, buildingID, err)
			continue
		}
		log.Debug("设备轮播同步成功 | 设备ID: %d | 建筑ID: %d | 操作: %s",
			device.ID, buildingID, action)
	}

	return nil
}

// updateDeviceNoticeCarousel 更新单个设备的通知轮播列表
func (s *NoticeSyncService) updateDeviceNoticeCarousel(ctx context.Context, deviceID uint, noticeMD5 string, noticeID uint, action string) error {
	_ = ctx       // 暂时未使用，保留用于未来扩展
	_ = noticeMD5 // 暂时未使用，保留用于未来扩展
	// 获取设备的当前轮播列表
	var device base_models.Device
	if err := s.db.Select("notice_carousel_list").First(&device, deviceID).Error; err != nil {
		return fmt.Errorf("failed to get device: %v", err)
	}

	// 解析当前的轮播列表
	var currentList []uint
	if device.NoticeCarouselList != nil {
		if err := json.Unmarshal(device.NoticeCarouselList, &currentList); err != nil {
			log.Warn("解析设备轮播列表失败 | 设备ID: %d | 错误: %v", deviceID, err)
			currentList = []uint{}
		}
	}

	var newList []uint
	switch action {
	case "add":
		// 检查通知是否已在列表中
		exists := false
		for _, id := range currentList {
			if id == noticeID {
				exists = true
				break
			}
		}
		if !exists {
			// 添加到列表末尾
			newList = append(currentList, noticeID)
			log.Debug("添加通知到设备轮播 | 设备ID: %d | 通知ID: %d", deviceID, noticeID)
		} else {
			// 已存在，无需更新
			return nil
		}

	case "remove":
		// 从列表中移除通知
		for _, id := range currentList {
			if id != noticeID {
				newList = append(newList, id)
			}
		}
		log.Debug("从设备轮播移除通知 | 设备ID: %d | 通知ID: %d", deviceID, noticeID)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	// 更新设备的轮播列表
	jsonData, err := json.Marshal(newList)
	if err != nil {
		return fmt.Errorf("failed to marshal notice carousel list: %v", err)
	}

	if err := s.db.Model(&base_models.Device{}).
		Where("id = ?", deviceID).
		Update("notice_carousel_list", jsonData).Error; err != nil {
		return fmt.Errorf("failed to update device notice carousel: %v", err)
	}

	return nil
}

// syncAllDeviceNoticeCarousels 同步所有设备的通知轮播列表
func (s *NoticeSyncService) syncAllDeviceNoticeCarousels(ctx context.Context, buildingID uint, addedNotices []int, removedNotices []int) error {
	log.Info("开始同步所有设备轮播列表 | 建筑ID: %d | 新增: %d | 删除: %d",
		buildingID, len(addedNotices), len(removedNotices))

	// 处理新增的通知
	for _, noticeID := range addedNotices {
		// 获取通知的MD5值
		var notice base_models.Notice
		if err := s.db.Preload("File").First(&notice, noticeID).Error; err != nil {
			log.Warn("获取通知信息失败 | 通知ID: %d | 错误: %v", noticeID, err)
			continue
		}

		if notice.File != nil {
			noticeMD5 := notice.File.Md5
			if err := s.syncDeviceNoticeCarousel(ctx, buildingID, noticeMD5, uint(noticeID), "add"); err != nil {
				log.Warn("同步设备轮播失败(新增) | 建筑ID: %d | 通知ID: %d | 错误: %v",
					buildingID, noticeID, err)
			}
		}
	}

	// 处理删除的通知
	for _, noticeID := range removedNotices {
		// 获取通知的MD5值
		var notice base_models.Notice
		if err := s.db.Preload("File").First(&notice, noticeID).Error; err != nil {
			log.Warn("获取通知信息失败 | 通知ID: %d | 错误: %v", noticeID, err)
			continue
		}

		if notice.File != nil {
			noticeMD5 := notice.File.Md5
			if err := s.syncDeviceNoticeCarousel(ctx, buildingID, noticeMD5, uint(noticeID), "remove"); err != nil {
				log.Warn("同步设备轮播失败(删除) | 建筑ID: %d | 通知ID: %d | 错误: %v",
					buildingID, noticeID, err)
			}
		}
	}

	log.Info("设备轮播同步完成 | 建筑ID: %d", buildingID)
	return nil
}

// 修改clearAllCaches函数，添加对同步MD5缓存的清除
func (s *NoticeSyncService) clearAllCaches(ctx context.Context) error {
	// 获取所有建筑物
	var buildings []base_models.Building
	if err := s.db.Find(&buildings).Error; err != nil {
		return fmt.Errorf("failed to get buildings: %v", err)
	}

	// 清除每个建筑物的缓存
	for _, building := range buildings {
		// 清除通知ID缓存
		idKey := fmt.Sprintf("building_notice_ids:%d", building.ID)
		if err := s.redis.Del(ctx, idKey).Err(); err != nil {
			log.Warn("警告: 清除通知ID缓存失败 | 建筑ID: %d | 错误: %v", building.ID, err)
		}

		// 清除通知数量缓存
		countKey := fmt.Sprintf("building_ismart_notice_count:%d", building.ID)
		if err := s.redis.Del(ctx, countKey).Err(); err != nil {
			log.Warn("警告: 清除通知数量缓存失败 | 建筑ID: %d | 错误: %v", building.ID, err)
		}

		// 清除建筑物缓存
		buildingKey := fmt.Sprintf("building:%d", building.ID)
		if err := s.redis.Del(ctx, buildingKey).Err(); err != nil {
			log.Warn("警告: 清除建筑物缓存失败 | 建筑ID: %d | 错误: %v", building.ID, err)
		}

		// 清除同步的通知MD5缓存
		syncedMD5Key := fmt.Sprintf("%s:%d", syncedNoticeMD5Prefix, building.ID)
		if err := s.redis.Del(ctx, syncedMD5Key).Err(); err != nil {
			log.Warn("警告: 清除同步MD5缓存失败 | 建筑ID: %d | 错误: %v", building.ID, err)
		}
	}

	log.Info("所有缓存已清除成功")
	return nil
}

// clearBuildingCaches 清除指定建筑物的所有缓存
func (s *NoticeSyncService) clearBuildingCaches(ctx context.Context, buildingID uint) error {
	// 清除通知ID缓存
	idKey := fmt.Sprintf("building_notice_ids:%d", buildingID)
	if err := s.redis.Del(ctx, idKey).Err(); err != nil {
		return fmt.Errorf("failed to clear notice IDs cache: %v", err)
	}

	// 清除通知数量缓存
	countKey := fmt.Sprintf("building_ismart_notice_count:%d", buildingID)
	if err := s.redis.Del(ctx, countKey).Err(); err != nil {
		return fmt.Errorf("failed to clear notice count cache: %v", err)
	}

	// 清除同步的MD5缓存
	syncedMD5Key := fmt.Sprintf("%s:%d", syncedNoticeMD5Prefix, buildingID)
	if err := s.redis.Del(ctx, syncedMD5Key).Err(); err != nil {
		return fmt.Errorf("failed to clear synced notice MD5s cache: %v", err)
	}

	// 清除同步的通知信息缓存
	syncedInfoKey := fmt.Sprintf("%s:%d", syncedNoticeIDsPrefix, buildingID)
	if err := s.redis.Del(ctx, syncedInfoKey).Err(); err != nil {
		return fmt.Errorf("failed to clear synced notice infos cache: %v", err)
	}

	// 清除建筑物信息缓存
	buildingKey := fmt.Sprintf("building:%d", buildingID)
	if err := s.redis.Del(ctx, buildingKey).Err(); err != nil {
		return fmt.Errorf("failed to clear building cache: %v", err)
	}

	log.Info("建筑物缓存清除完成 | 建筑ID: %d", buildingID)
	return nil
}

// 修改ManualSyncBuildingNotices函数
func (s *NoticeSyncService) ManualSyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error) {
	ctx := context.Background()
	log.Info("开始手动同步建筑物通知 | 建筑ID: %d", buildingID)

	// 清除该建筑物的所有缓存
	if err := s.clearBuildingCaches(ctx, buildingID); err != nil {
		log.Warn("警告: 清除建筑物缓存失败 | 建筑ID: %d | 错误: %v", buildingID, err)
	}

	// 执行同步
	result, err := s.SyncBuildingNotices(buildingID, claims)
	if err != nil {
		return nil, fmt.Errorf("手动同步失败: %v", err)
	}

	// 手动同步完成后，强制更新所有设备的轮播列表
	needSyncNotices := result["need_sync_notices"].([]int)
	changeNotices := result["change_notices"].([]int)

	if err := s.syncAllDeviceNoticeCarousels(ctx, buildingID, needSyncNotices, changeNotices); err != nil {
		log.Warn("警告: 手动同步后设备轮播更新失败 | 建筑ID: %d | 错误: %v", buildingID, err)
	} else {
		log.Info("手动同步后设备轮播更新完成 | 建筑ID: %d", buildingID)
	}

	log.Info("手动同步完成 | 建筑ID: %d", buildingID)
	return result, nil
}

// 10. StartSyncScheduler
func (s *NoticeSyncService) StartSyncScheduler(ctx context.Context) {
	syncInterval := getNoticeSyncInterval()
	ticker := time.NewTicker(syncInterval)
	countdownTicker := time.NewTicker(1 * time.Minute)
	log.Info("调度器已启动")

	// Clear all caches before starting
	if err := s.clearAllCaches(ctx); err != nil {
		log.Warn("警告: 启动时清除缓存失败 | 错误: %v", err)
	}

	go func() {
		// Create admin claims for sync
		adminClaims := jwt.MapClaims{
			"id":      float64(1),
			"isAdmin": true,
			"email":   "admin@example.com",
		}

		// 立即执行一次同步
		log.Info("正在运行初始同步...")
		s.runSync(adminClaims)
		nextSync := time.Now().Add(syncInterval)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				countdownTicker.Stop()
				log.Info("调度器已停止")
				return
			case <-countdownTicker.C:
				remaining := time.Until(nextSync).Round(time.Minute)
				log.Info("下次同步在 %d 分钟后", int(remaining.Minutes()))
			case <-ticker.C:
				log.Info("正在运行计划同步...")
				s.runSync(adminClaims)
				nextSync = time.Now().Add(syncInterval)
			}
		}
	}()
}

// 11. runSync
func (s *NoticeSyncService) runSync(adminClaims jwt.MapClaims) {
	// Get all buildings
	var buildings []base_models.Building
	if err := s.db.Find(&buildings).Error; err != nil {
		log.Error("获取建筑物失败 | 错误: %v", err)
		return
	}

	buildingCount := len(buildings)
	log.Info("找到 %d 个建筑物进行同步", buildingCount)

	// Get number of CPU cores
	cpuCores := runtime.NumCPU()

	// Calculate optimal number of building workers based on CPU cores
	maxBuildingWorkers := calculateBuildingWorkers(buildingCount, cpuCores)

	log.Info("系统有 %d 个CPU核心，使用 %d 个并发建筑工人", cpuCores, maxBuildingWorkers)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxBuildingWorkers)

	// Create a channel for collecting results
	resultChan := make(chan struct {
		buildingID uint
		name       string
		result     gin.H
		err        error
	}, buildingCount)

	// Start sync for each building concurrently
	for _, building := range buildings {
		wg.Add(1)
		go func(b base_models.Building) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Info("开始同步建筑物 %d (%s)", b.ID, b.Name)
			result, err := s.SyncBuildingNotices(b.ID, adminClaims)

			// Send result to channel
			resultChan <- struct {
				buildingID uint
				name       string
				result     gin.H
				err        error
			}{
				buildingID: b.ID,
				name:       b.Name,
				result:     result,
				err:        err,
			}
		}(building)
	}

	// Start a goroutine to close result channel when all syncs are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results as they come in
	for res := range resultChan {
		if res.err != nil {
			log.Error("建筑通知同步失败 | 建筑ID: %d | 名称: %s | 错误: %v",
				res.buildingID, res.name, res.err)
			continue
		}

		// Consolidated statistics output
		log.Info("建筑通知同步完成 | 建筑ID: %d | 名称: %s", res.buildingID, res.name)
		log.Info("同步统计 | 建筑ID: %d | 总处理: %d | 成功: %d | 已同步: %d | 失败: %d | 待删除: %d",
			res.buildingID, res.result["totalProcessed"], res.result["successCount"],
			res.result["hasSyncedCount"], len(res.result["failedNotices"].([]string)),
			res.result["deleteCount"])

		// 同步设备轮播列表
		needSync := res.result["need_sync_notices"].([]int)
		changes := res.result["change_notices"].([]int)

		if len(needSync) > 0 || len(changes) > 0 {
			ctx := context.Background()
			if err := s.syncAllDeviceNoticeCarousels(ctx, res.buildingID, needSync, changes); err != nil {
				log.Warn("定时同步后设备轮播更新失败 | 建筑ID: %d | 名称: %s | 错误: %v",
					res.buildingID, res.name, err)
			} else {
				log.Info("定时同步后设备轮播更新完成 | 建筑ID: %d | 名称: %s | 新增: %d | 删除: %d",
					res.buildingID, res.name, len(needSync), len(changes))
			}
		}

		// Only show detailed IDs if there are changes
		if len(needSync) > 0 {
			log.Debug("需要同步的通知IDs | 建筑ID: %d | IDs: %v", res.buildingID, needSync)
		}
		if len(changes) > 0 {
			log.Debug("变更的通知IDs | 建筑ID: %d | IDs: %v", res.buildingID, changes)
		}

		// Only show failed notices if there are any
		failedNotices := res.result["failedNotices"].([]string)
		if len(failedNotices) > 0 {
			log.Warn("同步失败的通知 | 建筑ID: %d", res.buildingID)
			for _, notice := range failedNotices {
				log.Warn("失败通知详情 | 建筑ID: %d | %s", res.buildingID, notice)
			}
		}
	}
}

// processNotice handles the processing of a single notice
func (s *NoticeSyncService) processNotice(buildingID uint, oldNotice OldSystemNotice, fileContent []byte, claims jwt.MapClaims) error {
	fileSize := len(fileContent)
	md5Hash := md5.Sum(fileContent)
	md5Str := hex.EncodeToString(md5Hash[:])
	mimeType := http.DetectContentType(fileContent)

	// Get uploader info from claims
	var uploaderID uint
	var uploaderType field.FileUploaderType
	var uploaderEmail string

	if isAdmin, ok := claims["isAdmin"].(bool); ok && isAdmin {
		if id, ok := claims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = field.UploaderTypeSuperAdmin
			if email, ok := claims["email"].(string); ok {
				uploaderEmail = email
			}
		}
	}

	if uploaderID == 0 {
		uploaderID = 1 // Use default system admin ID if not set
		uploaderType = field.UploaderTypeSuperAdmin
		uploaderEmail = "admin@example.com"
	}

	// Check if file exists
	var existingFile base_models.File
	var shouldUpload bool = true
	var fileForNotice *base_models.File

	err := s.db.Where("md5 = ?", md5Str).First(&existingFile).Error
	if err == nil {
		// File exists, check if it's associated with any notice
		var existingNotice base_models.Notice
		err := s.db.Where("file_id = ?", existingFile.ID).First(&existingNotice).Error

		if err == nil {
			// Check if notice is bound to current building
			var count int64
			err = s.db.Table("notice_buildings").
				Where("notice_id = ? AND building_id = ?", existingNotice.ID, buildingID).
				Count(&count).Error

			if err == nil && count > 0 {
				return nil
			}

			// Check if notice is bound to other buildings
			err = s.db.Table("notice_buildings").
				Where("notice_id = ?", existingNotice.ID).
				Count(&count).Error

			if err == nil {
				if count > 0 {
					shouldUpload = false
					fileForNotice = &existingFile
				} else {
					if err := s.db.Delete(&existingNotice).Error; err != nil {
						return fmt.Errorf("failed to delete old notice: %v", err)
					}
					shouldUpload = false
					fileForNotice = &existingFile
				}
			}
		} else {
			shouldUpload = false
			fileForNotice = &existingFile
		}
	} else if err == gorm.ErrRecordNotFound {
		shouldUpload = true

		// Generate unique file name using UUID
		uuid := uuid.New().String()
		fileName := uuid + ".pdf"
		// Use consistent directory structure with regular uploads
		dir := "iboard/pdf/"
		objectKey := dir + fileName

		// Get OSS configuration from environment variables
		host := os.Getenv("HOST")
		if host == "" {
			host = "http://idreamsky.oss-cn-beijing.aliyuncs.com"
		}

		// Create file record with OSS path
		fileForNotice = &base_models.File{
			Path:         host + "/" + objectKey, // Store the complete OSS URL
			Size:         int64(fileSize),
			MimeType:     mimeType,
			Oss:          "aliyun",
			UploaderType: uploaderType,
			UploaderID:   uploaderID,
			Uploader:     uploaderEmail,
			Md5:          md5Str,
		}

		if shouldUpload {
			// Get upload parameters from service
			uploadParams, err := s.uploadService.GetUploadParamsSync(objectKey)
			if err != nil {
				return fmt.Errorf("failed to get upload params: %v", err)
			}

			// Upload file to OSS
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			formFields := []struct {
				key      string
				paramKey string
			}{
				{key: "key", paramKey: "dir"},
				{key: "policy", paramKey: "policy"},
				{key: "OSSAccessKeyId", paramKey: "accessid"},
				{key: "success_action_status", paramKey: ""},
				{key: "callback", paramKey: "callback"},
				{key: "signature", paramKey: "signature"},
			}

			for _, field := range formFields {
				var fieldValue string
				switch field.key {
				case "key":
					fieldValue = objectKey
				case "success_action_status":
					fieldValue = "200"
				default:
					if field.paramKey != "" {
						if val, ok := uploadParams[field.paramKey]; ok {
							fieldValue = fmt.Sprintf("%v", val)
						} else {
							continue
						}
					}
				}
				if err := w.WriteField(field.key, fieldValue); err != nil {
					return fmt.Errorf("failed to write form field %s: %v", field.key, err)
				}
			}

			fw, err := w.CreateFormFile("file", objectKey)
			if err != nil {
				return fmt.Errorf("failed to create form file: %v", err)
			}
			if _, err = io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			w.Close()

			uploadReq, err := http.NewRequest("POST", uploadParams["host"].(string), &b)
			if err != nil {
				return fmt.Errorf("failed to create upload request: %v", err)
			}
			uploadReq.Header.Set("Content-Type", w.FormDataContentType())

			client := &http.Client{}
			uploadResp, err := client.Do(uploadReq)
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
			defer uploadResp.Body.Close()

			if uploadResp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(uploadResp.Body)
				return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(respBody))
			}
		}

		// Create file record
		if err := s.fileService.Create(fileForNotice); err != nil {
			return fmt.Errorf("failed to create file record: %v", err)
		}
	}

	// Create notice
	notice := &base_models.Notice{
		Title:          oldNotice.MessTitle,
		Description:    oldNotice.MessTitle,
		Type:           s.mapNoticeType(oldNotice.MessType),
		Status:         field.Status("active"),
		StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		EndTime:        time.Date(2100, 2, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		IsPublic:       true,
		IsIsmartNotice: true,
		FileID:         &fileForNotice.ID,
		FileType:       field.FileTypePdf,
	}

	// Create notice and bind to building in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return fmt.Errorf("failed to create notice: %v", err)
		}

		if err := tx.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)",
			notice.ID, buildingID).Error; err != nil {
			return fmt.Errorf("failed to bind notice to building: %v", err)
		}

		return nil
	})
}

// 9. mapNoticeType
func (s *NoticeSyncService) mapNoticeType(oldType string) field.NoticeType {
	switch oldType {
	case string(field.NoticeOldTypeCommon):
		return field.NoticeTypeNormal
	case string(field.NoticeOldTypeIo):
		return field.NoticeTypeBuilding
	case string(field.NoticeOldTypeUrgent):
		return field.NoticeTypeUrgent
	case string(field.NoticeOldTypeGovernment):
		return field.NoticeTypeGovernment
	default:
		return field.NoticeTypeNormal
	}
}
