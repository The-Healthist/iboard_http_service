package services

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/models"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminService interface {
	Create(admin *models.BuildingAdmin, buildingIDs []string, permissions []string) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.BuildingAdmin, models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}, buildingIDs []string, permissions []string) error
	Delete(ids []uint) error
}

type BuildingAdminService struct {
	db *gorm.DB
}

func NewBuildingAdminService(db *gorm.DB) InterfaceBuildingAdminService {
	return &BuildingAdminService{db: db}
}

func (s *BuildingAdminService) Create(admin *models.BuildingAdmin, buildingIDs []string, permissions []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(admin).Error; err != nil {
			return err
		}

		if len(buildingIDs) > 0 {
			var buildings []models.Building
			if err := tx.Where("id IN ?", buildingIDs).Find(&buildings).Error; err != nil {
				return err
			}
			if err := tx.Model(admin).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		if len(permissions) > 0 {
			var perms []models.Permission
			if err := tx.Where("name IN ?", permissions).Find(&perms).Error; err != nil {
				return err
			}
			if err := tx.Model(admin).Association("Permissions").Replace(perms); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BuildingAdminService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]models.BuildingAdmin, models.PaginationResult, error) {
	var admins []models.BuildingAdmin
	var total int64
	db := s.db.Model(&models.BuildingAdmin{}).
		Preload("Buildings").
		Preload("Permissions")

	if buildingID, ok := query["buildingId"].(string); ok && buildingID != "" {
		db = db.Where("building_id = ?", buildingID)
	}

	if status, ok := query["status"].(*bool); ok && status != nil {
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

func (s *BuildingAdminService) Update(id uint, updates map[string]interface{}, buildingIDs []string, permissions []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		admin := &models.BuildingAdmin{}
		if err := tx.First(admin, id).Error; err != nil {
			return err
		}

		if err := tx.Model(admin).Updates(updates).Error; err != nil {
			return err
		}

		if buildingIDs != nil {
			var buildings []models.Building
			if err := tx.Where("id IN ?", buildingIDs).Find(&buildings).Error; err != nil {
				return err
			}
			if err := tx.Model(admin).Association("Buildings").Replace(buildings); err != nil {
				return err
			}
		}

		if permissions != nil {
			var perms []models.Permission
			if err := tx.Where("name IN ?", permissions).Find(&perms).Error; err != nil {
				return err
			}
			if err := tx.Model(admin).Association("Permissions").Replace(perms); err != nil {
				return err
			}
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
