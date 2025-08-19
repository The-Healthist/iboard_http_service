package http_relationship_service

import (
	"encoding/json"
	"errors"
	"fmt"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InterfaceNoticeBuildingService interface {
	BindBuildings(noticeID uint, buildingIDs []uint) error
	UnbindBuildings(noticeID uint, buildingIDs []uint) error
	GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error)
	GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	NoticeExists(noticeID uint) (bool, error)
}

type NoticeBuildingService struct {
	db *gorm.DB
}

func NewNoticeBuildingService(db *gorm.DB) InterfaceNoticeBuildingService {
	log.Debug("初始化NoticeBuildingService")
	return &NoticeBuildingService{db: db}
}

func (s *NoticeBuildingService) BindBuildings(noticeID uint, buildingIDs []uint) error {
	log.Info("绑定通知到建筑 | 通知ID: %d | 建筑IDs: %v", noticeID, buildingIDs)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var notice base_models.Notice
		if err := tx.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
			log.Warn("通知不存在 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		var buildings []base_models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			log.Error("查询建筑失败 | 建筑IDs: %v | 错误: %v", buildingIDs, err)
			return err
		}

		if len(buildings) != len(buildingIDs) {
			log.Warn("部分建筑未找到 | 请求建筑数: %d | 找到建筑数: %d", len(buildingIDs), len(buildings))
		}

		// 检查现有关系
		var existingRelations []struct {
			BuildingID uint `gorm:"column:building_id"`
		}
		if err := tx.Table("notice_buildings").
			Select("building_id").
			Where("notice_id = ? AND building_id IN ?", noticeID, buildingIDs).
			Find(&existingRelations).Error; err != nil {
			log.Error("查询现有关系失败 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		// 过滤掉已存在的关系
		existingBuildingIDs := make(map[uint]bool)
		for _, relation := range existingRelations {
			existingBuildingIDs[relation.BuildingID] = true
		}

		var newBuildingIDs []uint
		for _, buildingID := range buildingIDs {
			if !existingBuildingIDs[buildingID] {
				newBuildingIDs = append(newBuildingIDs, buildingID)
			}
		}

		if len(newBuildingIDs) == 0 {
			log.Info("所有建筑都已绑定到该通知 | 通知ID: %d", noticeID)
			return nil
		}

		// 创建新的绑定关系
		for _, buildingID := range newBuildingIDs {
			if err := tx.Exec(
				"INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)",
				noticeID, buildingID,
			).Error; err != nil {
				log.Error("创建通知建筑关系失败 | 通知ID: %d | 建筑ID: %d | 错误: %v", noticeID, buildingID, err)
				return err
			}
		}

		// 同步更新相关 device 的轮播列表
		if err := s.syncDeviceNoticeCarouselLists(tx, noticeID, newBuildingIDs, true); err != nil {
			log.Error("同步设备通知轮播列表失败 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		log.Info("成功绑定通知到建筑 | 通知ID: %d | 新绑定建筑数量: %d", noticeID, len(newBuildingIDs))
		return nil
	})
}

func (s *NoticeBuildingService) UnbindBuildings(noticeID uint, buildingIDs []uint) error {
	log.Info("解绑通知与建筑 | 通知ID: %d | 建筑数量: %d", noticeID, len(buildingIDs))
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 同步更新相关 device 的轮播列表（在删除关系之前）
		if err := s.syncDeviceNoticeCarouselLists(tx, noticeID, buildingIDs, false); err != nil {
			log.Error("同步设备通知轮播列表失败 | 通知ID: %d | 错误: %v", noticeID, err)
			return err
		}

		// 删除绑定关系
		if err := tx.Exec(
			"DELETE FROM notice_buildings WHERE notice_id = ? AND building_id IN ?",
			noticeID, buildingIDs,
		).Error; err != nil {
			log.Error("删除通知建筑关系失败 | 通知ID: %d | 建筑IDs: %v | 错误: %v", noticeID, buildingIDs, err)
			return err
		}

		log.Info("成功解绑通知与建筑 | 通知ID: %d | 解绑建筑数量: %d", noticeID, len(buildingIDs))
		return nil
	})
}

func (s *NoticeBuildingService) GetBuildingsByNoticeID(noticeID uint) ([]base_models.Building, error) {
	log.Info("获取通知关联的建筑 | 通知ID: %d", noticeID)
	var notice base_models.Notice
	if err := s.db.Preload("Buildings").First(&notice, noticeID).Error; err != nil {
		log.Warn("通知不存在 | 通知ID: %d | 错误: %v", noticeID, err)
		return nil, err
	}
	if notice.Buildings == nil {
		log.Debug("通知未关联任何建筑 | 通知ID: %d", noticeID)
		return []base_models.Building{}, nil
	}
	log.Info("成功获取通知关联的建筑 | 通知ID: %d | 建筑数量: %d", noticeID, len(notice.Buildings))
	return notice.Buildings, nil
}

func (s *NoticeBuildingService) GetNoticesByBuildingID(buildingID uint) ([]base_models.Notice, error) {
	log.Info("获取建筑关联的通知 | 建筑ID: %d", buildingID)
	var building base_models.Building
	if err := s.db.
		Preload("Notices.File").
		Preload("Notices").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("建筑不存在 | 建筑ID: %d", buildingID)
			return []base_models.Notice{}, nil
		}
		log.Error("查询建筑失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return nil, err
	}
	log.Info("成功获取建筑关联的通知 | 建筑ID: %d | 通知数量: %d", buildingID, len(building.Notices))
	return building.Notices, nil
}

func (s *NoticeBuildingService) NoticeExists(noticeID uint) (bool, error) {
	log.Debug("检查通知是否存在 | 通知ID: %d", noticeID)
	var notice base_models.Notice
	err := s.db.First(&notice, noticeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("通知不存在 | 通知ID: %d", noticeID)
			return false, nil
		}
		log.Error("检查通知存在性失败 | 通知ID: %d | 错误: %v", noticeID, err)
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BuildingExists(buildingID uint) (bool, error) {
	log.Debug("检查建筑是否存在 | 建筑ID: %d", buildingID)
	var building base_models.Building
	err := s.db.First(&building, buildingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Debug("建筑不存在 | 建筑ID: %d", buildingID)
			return false, nil
		}
		log.Error("检查建筑存在性失败 | 建筑ID: %d | 错误: %v", buildingID, err)
		return false, err
	}
	return true, nil
}

func (s *NoticeBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	log.Debug("批量检查建筑存在性 | 建筑IDs: %v", buildingIDs)
	var count int64
	err := s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		log.Error("批量检查建筑存在性失败 | 错误: %v", err)
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		log.Debug("所有建筑都存在 | 建筑数量: %d", len(buildingIDs))
		return []uint{}, nil
	}

	var existingIDs []uint
	err = s.db.Model(&base_models.Building{}).Where("id IN ?", buildingIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		log.Error("获取存在的建筑ID失败 | 错误: %v", err)
		return nil, err
	}

	existingMap := make(map[uint]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	var missingIDs []uint
	for _, id := range buildingIDs {
		if !existingMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	log.Warn("部分建筑不存在 | 请求建筑数: %d | 缺失建筑数: %d | 缺失IDs: %v",
		len(buildingIDs), len(missingIDs), missingIDs)
	return missingIDs, nil
}

// syncDeviceNoticeCarouselLists 同步更新 device 的通知轮播列表
func (s *NoticeBuildingService) syncDeviceNoticeCarouselLists(tx *gorm.DB, noticeID uint, buildingIDs []uint, isBind bool) error {
	// 获取这些建筑下的所有 device
	var devices []base_models.Device
	if err := tx.Where("building_id IN ?", buildingIDs).Find(&devices).Error; err != nil {
		return fmt.Errorf("failed to get devices: %v", err)
	}

	if len(devices) == 0 {
		log.Info("没有找到需要同步的设备 | 建筑IDs: %v", buildingIDs)
		return nil
	}

	for _, device := range devices {
		if err := s.updateDeviceNoticeCarouselList(tx, &device, noticeID, isBind); err != nil {
			log.Error("更新设备通知轮播列表失败 | 设备ID: %d | 通知ID: %d | 错误: %v", device.ID, noticeID, err)
			return err
		}
	}

	log.Info("成功同步设备通知轮播列表 | 通知ID: %d | 设备数量: %d | 操作: %s", noticeID, len(devices), map[bool]string{true: "bind", false: "unbind"}[isBind])
	return nil
}

// updateDeviceNoticeCarouselList 更新单个设备的通知轮播列表
func (s *NoticeBuildingService) updateDeviceNoticeCarouselList(tx *gorm.DB, device *base_models.Device, noticeID uint, isBind bool) error {
	currentList, err := s.getCarouselListFromJSON(device.NoticeCarouselList)
	if err != nil {
		return fmt.Errorf("failed to parse notice carousel list: %v", err)
	}

	// 更新列表
	if isBind {
		// 绑定：添加到列表末尾（如果不存在）
		if !s.containsID(currentList, noticeID) {
			currentList = append(currentList, noticeID)
		}
	} else {
		// 解绑：从列表中移除
		currentList = s.removeID(currentList, noticeID)
	}

	// 保存更新后的列表
	jsonData, err := json.Marshal(currentList)
	if err != nil {
		return fmt.Errorf("failed to marshal notice carousel list: %v", err)
	}

	if err := tx.Model(&base_models.Device{}).Where("id = ?", device.ID).Update("notice_carousel_list", jsonData).Error; err != nil {
		return fmt.Errorf("failed to save notice carousel list: %v", err)
	}

	return nil
}

// getCarouselListFromJSON 从 JSON 字段解析轮播列表
func (s *NoticeBuildingService) getCarouselListFromJSON(jsonData datatypes.JSON) ([]uint, error) {
	if len(jsonData) == 0 {
		return []uint{}, nil
	}
	var list []uint
	if err := json.Unmarshal(jsonData, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// containsID 检查列表中是否包含指定 ID
func (s *NoticeBuildingService) containsID(list []uint, id uint) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}

// removeID 从列表中移除指定 ID
func (s *NoticeBuildingService) removeID(list []uint, id uint) []uint {
	var result []uint
	for _, item := range list {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}
