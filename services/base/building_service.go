package base_services

import (
	"errors"
	"fmt"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceBuildingService interface {
	Create(building *base_models.Building) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Building, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Building, error)
}

type BuildingService struct {
	db *gorm.DB
}

func NewBuildingService(db *gorm.DB) InterfaceBuildingService {
	return &BuildingService{db: db}
}

func (s *BuildingService) Create(building *models.Building) error {
	return s.db.Create(building).Error
}

func (s *BuildingService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.Building, models.PaginationResult, error) {
	var buildings []models.Building
	var total int64
	db := s.db.Model(&models.Building{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("name LIKE ? OR ismart_id LIKE ? OR remark LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
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

	if err := db.Preload("BuildingAdmins").Preload("Notices").Preload("Advertisements").
		Limit(pageSize).Offset(offset).Find(&buildings).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return buildings, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingService) Update(id uint, updates map[string]interface{}) error {
	result := s.db.Model(&models.Building{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("building not found")
	}
	return nil
}

func (s *BuildingService) Delete(ids []uint) error {
	result := s.db.Delete(&models.Building{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *BuildingService) GetByID(id uint) (*base_models.Building, error) {
	var building base_models.Building
	if err := s.db.Preload("BuildingAdmins").Preload("Notices").Preload("Advertisements").First(&building, id).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}
	return &building, nil
}
