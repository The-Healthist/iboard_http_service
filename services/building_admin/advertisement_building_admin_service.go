package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminAdvertisementService interface {
	Create(advertisement *base_models.Advertisement, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.Advertisement, error)
}

type BuildingAdminAdvertisementService struct {
	db                   *gorm.DB
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService
}

func NewBuildingAdminAdvertisementService(
	db *gorm.DB,
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService,
) InterfaceBuildingAdminAdvertisementService {
	return &BuildingAdminAdvertisementService{
		db:                   db,
		buildingAdminService: buildingAdminService,
	}
}

func (s *BuildingAdminAdvertisementService) Create(advertisement *base_models.Advertisement, email string) error {
	// 检查管理员��否有权限
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}

		// 自动关联到管理员的所有建筑物
		if err := tx.Model(advertisement).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminAdvertisementService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error) {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, b := range buildings {
		buildingIDs = append(buildingIDs, b.ID)
	}

	var advertisements []base_models.Advertisement
	var total int64
	db := s.db.Model(&base_models.Advertisement{}).
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Where("advertisement_buildings.building_id IN ?", buildingIDs).
		Group("advertisements.id")

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if err := db.Preload("File").
		Limit(pageSize).
		Offset(offset).
		Find(&advertisements).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return advertisements, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminAdvertisementService) Update(id uint, email string, updates map[string]interface{}) error {
	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return err
	} else if !exists {
		return errors.New("advertisement not found or no permission")
	}

	result := s.db.Model(&base_models.Advertisement{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *BuildingAdminAdvertisementService) Delete(id uint, email string) error {
	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return err
	} else if !exists {
		return errors.New("advertisement not found or no permission")
	}

	return s.db.Delete(&base_models.Advertisement{}, id).Error
}

func (s *BuildingAdminAdvertisementService) GetByID(id uint, email string) (*base_models.Advertisement, error) {
	var advertisement base_models.Advertisement

	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New("advertisement not found or no permission")
	}

	if err := s.db.Preload("File").First(&advertisement, id).Error; err != nil {
		return nil, err
	}
	return &advertisement, nil
}

func (s *BuildingAdminAdvertisementService) checkPermission(advertisementID uint, email string) (bool, error) {
	var count int64
	err := s.db.Model(&base_models.Advertisement{}).
		Joins("JOIN advertisement_buildings ON advertisements.id = advertisement_buildings.advertisement_id").
		Joins("JOIN building_admin_buildings ON advertisement_buildings.building_id = building_admin_buildings.building_id").
		Joins("JOIN building_admins ON building_admin_buildings.building_admin_id = building_admins.id").
		Where("advertisements.id = ? AND building_admins.email = ?", advertisementID, email).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
