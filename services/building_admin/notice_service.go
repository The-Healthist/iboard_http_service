package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminNoticeService interface {
	Create(notice *base_models.Notice, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.Notice, error)
}

type BuildingAdminNoticeService struct {
	db                   *gorm.DB
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService
}

func NewBuildingAdminNoticeService(
	db *gorm.DB,
	buildingAdminService relationship_service.InterfaceBuildingAdminBuildingService,
) InterfaceBuildingAdminNoticeService {
	return &BuildingAdminNoticeService{
		db:                   db,
		buildingAdminService: buildingAdminService,
	}
}

func (s *BuildingAdminNoticeService) Create(notice *base_models.Notice, email string) error {
	// 检查管理员是否有权限
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}

		// 自动关联到管理员的所有建筑物
		if err := tx.Model(notice).Association("Buildings").Append(&buildings); err != nil {
			return err
		}

		return nil
	})
}

func (s *BuildingAdminNoticeService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Notice, base_models.PaginationResult, error) {
	// 获取管理员的建筑物
	buildings, err := s.buildingAdminService.GetBuildingsByAdminEmail(email)
	if err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	var buildingIDs []uint
	for _, b := range buildings {
		buildingIDs = append(buildingIDs, b.ID)
	}

	var notices []base_models.Notice
	var total int64
	db := s.db.Model(&base_models.Notice{}).
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id IN ?", buildingIDs).
		Group("notices.id")

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if err := db.Preload("File").
		Limit(pageSize).
		Offset(offset).
		Find(&notices).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return notices, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminNoticeService) Update(id uint, email string, updates map[string]interface{}) error {
	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return err
	} else if !exists {
		return errors.New("notice not found or no permission")
	}

	result := s.db.Model(&base_models.Notice{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *BuildingAdminNoticeService) Delete(id uint, email string) error {
	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return err
	} else if !exists {
		return errors.New("notice not found or no permission")
	}

	return s.db.Delete(&base_models.Notice{}, id).Error
}

func (s *BuildingAdminNoticeService) GetByID(id uint, email string) (*base_models.Notice, error) {
	var notice base_models.Notice

	// 检查权限
	if exists, err := s.checkPermission(id, email); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.New("notice not found or no permission")
	}

	if err := s.db.Preload("File").First(&notice, id).Error; err != nil {
		return nil, err
	}
	return &notice, nil
}

// 检查管理员是否有权限操作该通知
func (s *BuildingAdminNoticeService) checkPermission(noticeID uint, email string) (bool, error) {
	var count int64
	err := s.db.Model(&base_models.Notice{}).
		Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Joins("JOIN building_admin_buildings ON notice_buildings.building_id = building_admin_buildings.building_id").
		Joins("JOIN building_admins ON building_admin_buildings.building_admin_id = building_admins.id").
		Where("notices.id = ? AND building_admins.email = ?", noticeID, email).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
