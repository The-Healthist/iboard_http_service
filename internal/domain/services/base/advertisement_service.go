package base_services

import (
	"encoding/json"
	"errors"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"gorm.io/gorm"
)

type InterfaceAdvertisementService interface {
	Create(advertisement *base_models.Advertisement) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error)
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Advertisement, error)
}

type AdvertisementService struct {
	db *gorm.DB
}

func NewAdvertisementService(db *gorm.DB) InterfaceAdvertisementService {
	if db == nil {
		panic("database connection is nil")
	}
	return &AdvertisementService{db: db}
}

func (s *AdvertisementService) Create(advertisement *base_models.Advertisement) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *AdvertisementService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error) {
	if s.db == nil {
		return nil, base_models.PaginationResult{}, errors.New("database connection is nil")
	}

	var advertisements []base_models.Advertisement
	var total int64
	db := s.db.Model(&base_models.Advertisement{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if advertisementType, ok := query["type"].(string); ok && advertisementType != "" {
		db = db.Where("type = ?", advertisementType)
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

	if err := db.Preload("File").
		Limit(pageSize).Offset(offset).
		Find(&advertisements).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	now := time.Now()
	for i := range advertisements {
		if advertisements[i].StartTime.IsZero() {
			advertisements[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if advertisements[i].EndTime.IsZero() {
			advertisements[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if advertisements[i].Status == "" {
			advertisements[i].Status = field.Status("active")
		}

		// Check if advertisement has expired
		if advertisements[i].EndTime.Before(now) && advertisements[i].Status == field.Status("active") {
			// Update status in database
			if err := s.db.Model(&base_models.Advertisement{}).Where("id = ?", advertisements[i].ID).Update("status", field.Status("inactive")).Error; err != nil {
				return nil, base_models.PaginationResult{}, err
			}
			// Update status in memory
			advertisements[i].Status = field.Status("inactive")
		}
	}

	return advertisements, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *AdvertisementService) Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var advertisement base_models.Advertisement

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先获取原始广告信息
		var originalAd base_models.Advertisement
		if err := tx.First(&originalAd, id).Error; err != nil {
			return err
		}

		// 更新广告
		if err := tx.Model(&base_models.Advertisement{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		// 重新获取更新后的广告信息
		if err := tx.Preload("File").Preload("Buildings").First(&advertisement, id).Error; err != nil {
			return err
		}

		// 检查是否需要同步更新设备轮播列表
		needSync := false
		var typeChanged, statusChanged bool
		var newType, newStatus interface{}
		var oldType = string(originalAd.Display)
		var oldStatus = string(originalAd.Status)

		if display, exists := updates["display"]; exists {
			typeChanged = true
			newType = display
			needSync = true
		}

		if status, exists := updates["status"]; exists {
			statusChanged = true
			newStatus = status
			needSync = true
		}

		// 如果需要同步，更新所有设备的轮播列表
		if needSync {
			if err := s.syncDeviceCarouselLists(tx, id, typeChanged, statusChanged, oldType, newType, oldStatus, newStatus); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &advertisement, nil
}

// syncDeviceCarouselLists 同步设备轮播列表
func (s *AdvertisementService) syncDeviceCarouselLists(tx *gorm.DB, adID uint, typeChanged, statusChanged bool, oldType interface{}, newType interface{}, oldStatus interface{}, newStatus interface{}) error {
	// 获取与该广告关联的建筑物ID
	var buildingIDs []uint
	if err := tx.Table("advertisement_buildings").Where("advertisement_id = ?", adID).Pluck("building_id", &buildingIDs).Error; err != nil {
		return err
	}

	// 如果没有关联的建筑物，直接返回
	if len(buildingIDs) == 0 {
		return nil
	}

	// 获取当前广告的最新信息
	var currentAd base_models.Advertisement
	if err := tx.First(&currentAd, adID).Error; err != nil {
		return err
	}

	// 获取这些建筑物下的所有设备
	var devices []base_models.Device
	if err := tx.Where("building_id IN ?", buildingIDs).Find(&devices).Error; err != nil {
		return err
	}

	for _, device := range devices {
		needUpdate := false

		// 处理TopAdvertisementCarouselList
		var topList []uint
		if device.TopAdvertisementCarouselList != nil {
			if err := json.Unmarshal(device.TopAdvertisementCarouselList, &topList); err == nil {
				originalTopList := make([]uint, len(topList))
				copy(originalTopList, topList)

				// 处理状态变化
				if statusChanged {
					if newStatus == field.Status("inactive") {
						// 状态变为inactive，从列表中移除
						topList = s.removeFromList(topList, adID)
						needUpdate = true
					} else if newStatus == field.Status("active") && oldStatus == string(field.Status("inactive")) {
						// 状态从inactive变为active，根据当前类型添加到相应列表
						currentDisplay := currentAd.Display

						// 如果当前类型是top或topfull，且不在列表中，则添加
						if (currentDisplay == field.AdvertisementDisplay("top") || currentDisplay == field.AdvertisementDisplay("topfull")) && !s.containsInList(topList, adID) {
							topList = append(topList, adID)
							needUpdate = true
						}
					}
				}

				// 如果类型变化，需要调整列表
				if typeChanged {
					// 如果新类型不是top或topfull，从top列表中移除
					if newType != field.AdvertisementDisplay("top") && newType != field.AdvertisementDisplay("topfull") {
						topList = s.removeFromList(topList, adID)
						needUpdate = true
					} else if oldType != "top" && oldType != "topfull" {
						// 如果旧类型不是top或topfull，新类型是，则添加到列表末尾
						if !s.containsInList(topList, adID) {
							topList = append(topList, adID)
							needUpdate = true
						}
					}
				}

				// 更新设备的TopAdvertisementCarouselList
				if needUpdate && !s.slicesEqual(originalTopList, topList) {
					topListBytes, _ := json.Marshal(topList)
					if err := tx.Model(&base_models.Device{}).Where("id = ?", device.ID).Update("top_advertisement_carousel_list", topListBytes).Error; err != nil {
						return err
					}
				}
			}
		}

		// 处理FullAdvertisementCarouselList
		var fullList []uint
		needUpdateFull := false
		if device.FullAdvertisementCarouselList != nil {
			if err := json.Unmarshal(device.FullAdvertisementCarouselList, &fullList); err == nil {
				originalFullList := make([]uint, len(fullList))
				copy(originalFullList, fullList)

				// 处理状态变化
				if statusChanged {
					if newStatus == field.Status("inactive") {
						// 状态变为inactive，从列表中移除
						fullList = s.removeFromList(fullList, adID)
						needUpdateFull = true
					} else if newStatus == field.Status("active") && oldStatus == string(field.Status("inactive")) {
						// 状态从inactive变为active，根据当前类型添加到相应列表
						currentDisplay := currentAd.Display

						// 如果当前类型是full或topfull，且不在列表中，则添加
						if (currentDisplay == field.AdvertisementDisplay("full") || currentDisplay == field.AdvertisementDisplay("topfull")) && !s.containsInList(fullList, adID) {
							fullList = append(fullList, adID)
							needUpdateFull = true
						}
					}
				}

				// 如果类型变化，需要调整列表
				if typeChanged {
					// 如果新类型不是full或topfull，从full列表中移除
					if newType != field.AdvertisementDisplay("full") && newType != field.AdvertisementDisplay("topfull") {
						fullList = s.removeFromList(fullList, adID)
						needUpdateFull = true
					} else if oldType != "full" && oldType != "topfull" {
						// 如果旧类型不是full或topfull，新类型是，则添加到列表末尾
						if !s.containsInList(fullList, adID) {
							fullList = append(fullList, adID)
							needUpdateFull = true
						}
					}
				}

				// 更新设备的FullAdvertisementCarouselList
				if needUpdateFull && !s.slicesEqual(originalFullList, fullList) {
					fullListBytes, _ := json.Marshal(fullList)
					if err := tx.Model(&base_models.Device{}).Where("id = ?", device.ID).Update("full_advertisement_carousel_list", fullListBytes).Error; err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
} // removeFromList 从列表中移除指定ID
func (s *AdvertisementService) removeFromList(list []uint, id uint) []uint {
	result := make([]uint, 0, len(list))
	for _, item := range list {
		if item != id {
			result = append(result, item)
		}
	}
	return result
}

// containsInList 检查列表中是否包含指定ID
func (s *AdvertisementService) containsInList(list []uint, id uint) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}

// slicesEqual 比较两个uint切片是否相等
func (s *AdvertisementService) slicesEqual(a, b []uint) bool {
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

func (s *AdvertisementService) Delete(ids []uint) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	result := s.db.Delete(&base_models.Advertisement{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *AdvertisementService) GetByID(id uint) (*base_models.Advertisement, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var advertisement base_models.Advertisement
	if err := s.db.Preload("File").First(&advertisement, id).Error; err != nil {
		return nil, err
	}

	if advertisement.StartTime.IsZero() {
		advertisement.StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if advertisement.EndTime.IsZero() {
		advertisement.EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if advertisement.Status == "" {
		advertisement.Status = field.Status("active")
	}
	if advertisement.File != nil && advertisement.File.ID == 0 {
		advertisement.File = nil
	}

	return &advertisement, nil
}
