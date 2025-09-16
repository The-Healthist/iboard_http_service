package base_services

import (
	"encoding/json"
	"errors"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"gorm.io/gorm"
)

type InterfaceNoticeService interface {
	Create(notice *base_models.Notice) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Notice, error)
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Notice, error)
}

type NoticeService struct {
	db *gorm.DB
}

func NewNoticeService(db *gorm.DB) InterfaceNoticeService {
	return &NoticeService{db: db}
}

func (s *NoticeService) Create(notice *base_models.Notice) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *NoticeService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error) {
	var notices []base_models.Notice
	var total int64
	db := s.db.Model(&base_models.Notice{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if noticeType, ok := query["type"].(string); ok && noticeType != "" {
		db = db.Where("type = ?", noticeType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Select("notices.*, notices.is_ismart_notice as is_ismart_notice").
		Limit(pageSize).Offset(offset).
		Find(&notices).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	now := time.Now()
	for i := range notices {
		if notices[i].StartTime.IsZero() {
			notices[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if notices[i].EndTime.IsZero() {
			notices[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if notices[i].Status == "" {
			notices[i].Status = field.Status("active")
		}
		if notices[i].FileType == "" {
			notices[i].FileType = field.FileTypePdf
		}

		// Check if notice has expired
		if notices[i].EndTime.Before(now) && notices[i].Status == field.Status("active") {
			// Update status in database
			if err := s.db.Model(&base_models.Notice{}).Where("id = ?", notices[i].ID).Update("status", field.Status("inactive")).Error; err != nil {
				return nil, base_models.PaginationResult{}, err
			}
			// Update status in memory
			notices[i].Status = field.Status("inactive")
		}
	}

	return notices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *NoticeService) Update(id uint, updates map[string]interface{}) (*base_models.Notice, error) {
	var notice base_models.Notice

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先获取原始通知信息
		var originalNotice base_models.Notice
		if err := tx.First(&originalNotice, id).Error; err != nil {
			return err
		}

		// 更新通知
		if err := tx.Model(&base_models.Notice{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		// 重新获取更新后的通知信息
		if err := tx.Preload("File").Preload("Buildings").First(&notice, id).Error; err != nil {
			return err
		}

		// 检查是否需要同步更新设备轮播列表
		needSync := false
		var statusChanged bool
		var newStatus, oldStatus interface{}

		if status, exists := updates["status"]; exists {
			statusChanged = true
			newStatus = status
			oldStatus = string(originalNotice.Status)
			needSync = true
		}

		// 如果需要同步，更新关联设备的轮播列表
		if needSync {
			if err := s.syncDeviceNoticeCarouselLists(tx, id, statusChanged, oldStatus, newStatus); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &notice, nil
}

func (s *NoticeService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.Notice{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *NoticeService) GetByID(id uint) (*base_models.Notice, error) {
	var notice base_models.Notice
	if err := s.db.Select("notices.*, notices.is_ismart_notice as is_ismart_notice").
		Preload("File").First(&notice, id).Error; err != nil {
		return nil, err
	}

	if notice.StartTime.IsZero() {
		notice.StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if notice.EndTime.IsZero() {
		notice.EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if notice.Status == "" {
		notice.Status = field.Status("active")
	}
	if notice.FileType == "" {
		notice.FileType = field.FileTypePdf
	}
	if notice.File != nil && notice.File.ID == 0 {
		notice.File = nil
	}

	return &notice, nil
}

// syncDeviceNoticeCarouselLists 同步设备的通知轮播列表
func (s *NoticeService) syncDeviceNoticeCarouselLists(tx *gorm.DB, noticeID uint, statusChanged bool, oldStatus interface{}, newStatus interface{}) error {
	// 获取与该通知关联的建筑物ID
	var buildingIDs []uint
	if err := tx.Table("notice_buildings").Where("notice_id = ?", noticeID).Pluck("building_id", &buildingIDs).Error; err != nil {
		return err
	}

	// 如果没有关联的建筑物，直接返回
	if len(buildingIDs) == 0 {
		return nil
	}

	// 获取这些建筑物下的所有设备
	var devices []base_models.Device
	if err := tx.Where("building_id IN ?", buildingIDs).Find(&devices).Error; err != nil {
		return err
	}

	for _, device := range devices {
		// 处理NoticeCarouselList
		var noticeList []uint
		needUpdate := false

		if device.NoticeCarouselList != nil {
			if err := json.Unmarshal(device.NoticeCarouselList, &noticeList); err == nil {
				originalNoticeList := make([]uint, len(noticeList))
				copy(originalNoticeList, noticeList)

				// 如果状态变为inactive，从列表中移除
				if statusChanged && newStatus == field.Status("inactive") {
					noticeList = s.removeNoticeFromList(noticeList, noticeID)
					needUpdate = true
				} else if statusChanged && newStatus == field.Status("active") && oldStatus == field.Status("inactive") {
					// 如果状态从inactive变为active，添加到列表末尾（如果不存在）
					if !s.containsNoticeInList(noticeList, noticeID) {
						noticeList = append(noticeList, noticeID)
						needUpdate = true
					}
				}

				// 更新设备的NoticeCarouselList
				if needUpdate && !s.noticeSlicesEqual(originalNoticeList, noticeList) {
					noticeListBytes, _ := json.Marshal(noticeList)
					if err := tx.Model(&base_models.Device{}).Where("id = ?", device.ID).Update("notice_carousel_list", noticeListBytes).Error; err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// removeNoticeFromList 从列表中移除指定通知ID
func (s *NoticeService) removeNoticeFromList(list []uint, id uint) []uint {
	result := make([]uint, 0, len(list))
	for _, item := range list {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}

// containsNoticeInList 检查列表中是否包含指定通知ID
func (s *NoticeService) containsNoticeInList(list []uint, id uint) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}

// noticeSlicesEqual 比较两个uint切片是否相等
func (s *NoticeService) noticeSlicesEqual(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
