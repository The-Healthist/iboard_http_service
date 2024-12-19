package base_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceAdvertisementService interface {
	Create(advertisement *base_models.Advertisement) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error)
	Delete(ids []uint) error
}

type AdvertisementService struct {
	db *gorm.DB
}

func NewAdvertisementService(db *gorm.DB) InterfaceAdvertisementService {
	return &AdvertisementService{db: db}
}

func (s *AdvertisementService) Create(advertisement *base_models.Advertisement) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *AdvertisementService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error) {
	var advertisements []base_models.Advertisement
	var total int64
	db := s.db.Model(&base_models.Advertisement{}).Preload("File").Preload("Buildings")

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

	if err := db.Limit(pageSize).Offset(offset).Find(&advertisements).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return advertisements, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *AdvertisementService) Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error) {
	var advertisement base_models.Advertisement

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&base_models.Advertisement{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Preload("File").Preload("Buildings").First(&advertisement, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &advertisement, nil
}

func (s *AdvertisementService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.Advertisement{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}
