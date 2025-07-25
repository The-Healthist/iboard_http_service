package base_services

import (
	"errors"
	"fmt"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminService interface {
	Create(admin *models.BuildingAdmin) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.BuildingAdmin, models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(ids []uint) error
	GetByEmail(email string) (*models.BuildingAdmin, error)
	ValidatePassword(admin *models.BuildingAdmin, password string) bool
	GetByID(id uint) (*models.BuildingAdmin, error)
}

type BuildingAdminService struct {
	db *gorm.DB
}

func NewBuildingAdminService(db *gorm.DB) InterfaceBuildingAdminService {
	return &BuildingAdminService{db: db}
}

func (s *BuildingAdminService) Create(admin *models.BuildingAdmin) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(admin).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *BuildingAdminService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.BuildingAdmin, models.PaginationResult, error) {
	var admins []models.BuildingAdmin
	var total int64
	db := s.db.Model(&models.BuildingAdmin{}).
		Preload("Buildings")

	if buildingID, ok := query["buildingId"].(string); ok && buildingID != "" {
		db = db.Where("building_id = ?", buildingID)
	}

	if status, ok := query["status"].(*field.Status); ok && status != nil {
		db = db.Where("status = ?", *status)
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

	if err := db.Limit(pageSize).Offset(offset).Find(&admins).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return admins, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminService) Update(id uint, updates map[string]interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		admin := &models.BuildingAdmin{}
		if err := tx.First(admin, id).Error; err != nil {
			return err
		}

		if err := tx.Model(admin).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *BuildingAdminService) Delete(ids []uint) error {
	result := s.db.Delete(&models.BuildingAdmin{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *BuildingAdminService) GetByEmail(email string) (*models.BuildingAdmin, error) {
	var admin models.BuildingAdmin
	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *BuildingAdminService) ValidatePassword(admin *models.BuildingAdmin, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	return err == nil
}

func (s *BuildingAdminService) GetByID(id uint) (*models.BuildingAdmin, error) {
	var admin models.BuildingAdmin
	if err := s.db.Preload("Buildings").First(&admin, id).Error; err != nil {
		return nil, fmt.Errorf("building admin not found: %v", err)
	}
	return &admin, nil
}
