package http_relationship_service

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"gorm.io/gorm"
)

// 自定义响应结构
type ServiceResponse struct {
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type InterfaceBuildingAdminBuildingService interface {
	BindBuildings(adminID uint, buildingIDs []uint) error
	UnbindBuildings(adminID uint, buildingIDs []uint) error
	GetBuildingsByAdminID(adminID uint) ([]models.Building, error)
	GetAdminsByBuildingID(buildingID uint) ([]models.BuildingAdmin, error)
	BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error)
	BuildingExists(buildingID uint) (bool, error)
	BuildingAdminExists(adminID uint) (bool, error)
	GetBuildingsByAdminEmail(email string) ([]models.Building, error)
}

type BuildingAdminBuildingService struct {
	db *gorm.DB
}

func NewBuildingAdminBuildingService(db *gorm.DB) InterfaceBuildingAdminBuildingService {
	return &BuildingAdminBuildingService{db: db}
}

func (s *BuildingAdminBuildingService) BindBuildings(adminID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var admin models.BuildingAdmin
		if err := tx.Preload("Buildings").First(&admin, adminID).Error; err != nil {
			return err
		}

		var buildings []models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&admin).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminBuildingService) UnbindBuildings(adminID uint, buildingIDs []uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var admin models.BuildingAdmin
		if err := tx.Preload("Buildings").First(&admin, adminID).Error; err != nil {
			return err
		}

		var buildings []models.Building
		if err := tx.Find(&buildings, buildingIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&admin).Association("Buildings").Delete(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminBuildingService) GetBuildingsByAdminID(adminID uint) ([]models.Building, error) {
	var admin models.BuildingAdmin
	if err := s.db.Preload("Buildings").First(&admin, adminID).Error; err != nil {
		return nil, err
	}
	if admin.Buildings == nil {
		return []models.Building{}, nil
	}
	return admin.Buildings, nil
}

func (s *BuildingAdminBuildingService) GetAdminsByBuildingID(buildingID uint) ([]models.BuildingAdmin, error) {
	var building models.Building
	if err := s.db.
		Preload("BuildingAdmins").
		First(&building, buildingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.BuildingAdmin{}, nil
		}
		return nil, err
	}
	return building.BuildingAdmins, nil
}

func (s *BuildingAdminBuildingService) BuildingAdminExists(adminID uint) (bool, error) {
	var admin models.BuildingAdmin
	err := s.db.First(&admin, adminID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *BuildingAdminBuildingService) BulkCheckBuildingsExistence(buildingIDs []uint) ([]uint, error) {
	var count int64
	err := s.db.Model(&models.Building{}).Where("id IN ?", buildingIDs).Count(&count).Error
	if err != nil {
		return nil, err
	}
	if int(count) == len(buildingIDs) {
		return []uint{}, nil
	}

	// 找出缺失的 Building IDs
	var existingIDs []uint
	err = s.db.Model(&models.Building{}).Where("id IN ?", buildingIDs).Pluck("id", &existingIDs).Error
	if err != nil {
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

	return missingIDs, nil
}

func (s *BuildingAdminBuildingService) BuildingExists(buildingID uint) (bool, error) {
	var building models.Building
	err := s.db.First(&building, buildingID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *BuildingAdminBuildingService) GetBuildingsByAdminEmail(email string) ([]models.Building, error) {
	var admin models.BuildingAdmin
	if err := s.db.Where("email = ?", email).Preload("Buildings").First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.Building{}, nil
		}
		return nil, err
	}
	if admin.Buildings == nil {
		return []models.Building{}, nil
	}
	return admin.Buildings, nil
}
