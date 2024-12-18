package services

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/models"
	"gorm.io/gorm"
)

type InterfaceAdvertisementService interface {
	Create(advertisement *models.Advertisement, buildingIDs []uint) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Advertisement, models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}, buildingIDs []uint) error
	Delete(ids []uint) error
}

type AdvertisementService struct {
	db *gorm.DB
}

func NewAdvertisementService(db *gorm.DB) InterfaceAdvertisementService {
	return &AdvertisementService{db: db}
}

func (s *AdvertisementService) Create(advertisement *models.Advertisement, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}

		if len(buildingIDs) > 0 {
			var buildings []models.Building
			if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
				return err
			}
			if err := tx.Model(advertisement).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *AdvertisementService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Advertisement, models.PaginationResult, error) {
	var advertisements []models.Advertisement
	var total int64
	db := s.db.Model(&models.Advertisement{}).Preload("File").Preload("Buildings")

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if adType, ok := query["type"].(string); ok && adType != "" {
		db = db.Where("type = ?", adType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Limit(pageSize).Offset(offset).Find(&advertisements).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return advertisements, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *AdvertisementService) Update(id uint, updates map[string]interface{}, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		advertisement := &models.Advertisement{}
		if err := tx.First(advertisement, id).Error; err != nil {
			return err
		}

		if err := tx.Model(advertisement).Updates(updates).Error; err != nil {
			return err
		}

		if buildingIDs != nil {
			var buildings []models.Building
			if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
				return err
			}
			if err := tx.Model(advertisement).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *AdvertisementService) Delete(ids []uint) error {
	result := s.db.Delete(&models.Advertisement{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}
