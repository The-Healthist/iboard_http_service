package base_services

import (
	"errors"
	"fmt"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

type InterfaceBuildingService interface {
	Create(building *base_models.Building) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Building, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Building, error)
	GetByCredentials(ismartID string, password string) (*base_models.Building, error)
	GetBuildingAdvertisements(buildingId uint) ([]base_models.Advertisement, error)
	GetBuildingNotices(buildingId uint) ([]base_models.Notice, error)
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

	if err := db.
		Preload("BuildingAdmins").
		Preload("BuildingAdmins.Buildings").
		Preload("Notices").
		Preload("Notices.File").
		Preload("Notices.Buildings").
		Preload("Advertisements").
		Preload("Advertisements.File").
		Preload("Advertisements.Buildings").
		Limit(pageSize).Offset(offset).
		Find(&buildings).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	for i := range buildings {
		if buildings[i].BuildingAdmins == nil {
			buildings[i].BuildingAdmins = []base_models.BuildingAdmin{}
		}
		if buildings[i].Notices == nil {
			buildings[i].Notices = []base_models.Notice{}
		} else {
			for j := range buildings[i].Notices {
				if buildings[i].Notices[j].Buildings == nil {
					buildings[i].Notices[j].Buildings = []base_models.Building{}
				}
			}
		}
		if buildings[i].Advertisements == nil {
			buildings[i].Advertisements = []base_models.Advertisement{}
		} else {
			for j := range buildings[i].Advertisements {
				if buildings[i].Advertisements[j].Buildings == nil {
					buildings[i].Advertisements[j].Buildings = []base_models.Building{}
				}
			}
		}
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
	if err := s.db.
		Preload("Notices", "is_public = ?", true).
		Preload("Notices.File").
		Preload("Advertisements", "is_public = ? AND active = ?", true, true).
		Preload("Advertisements.File").
		First(&building, id).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}

	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}
	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}

	for i := range building.Notices {
		if building.Notices[i].File.ID == 0 && building.Notices[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Notices[i].FileID).Error; err == nil {
				building.Notices[i].File = file
			}
		}
	}

	for i := range building.Advertisements {
		if building.Advertisements[i].File.ID == 0 && building.Advertisements[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Advertisements[i].FileID).Error; err == nil {
				building.Advertisements[i].File = file
			}
		}
	}

	return &building, nil
}

func (s *BuildingService) GetByCredentials(ismartID string, password string) (*base_models.Building, error) {
	var building base_models.Building
	if err := s.db.Where("ismart_id = ? AND password = ?", ismartID, password).
		Preload("Notices").
		Preload("Notices.File").
		Preload("Advertisements", " active = ?", true).
		Preload("Advertisements.File").
		First(&building).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials: %v", err)
	}

	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}
	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}

	for i := range building.Notices {
		if building.Notices[i].File.ID == 0 && building.Notices[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Notices[i].FileID).Error; err == nil {
				building.Notices[i].File = file
			}
		}
	}

	for i := range building.Advertisements {
		if building.Advertisements[i].File.ID == 0 && building.Advertisements[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Advertisements[i].FileID).Error; err == nil {
				building.Advertisements[i].File = file
			}
		}
	}

	building.Password = ""
	return &building, nil
}

func (s *BuildingService) GetBuildingAdvertisements(buildingId uint) ([]base_models.Advertisement, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Advertisements", "is_public = ? AND active = ?", true, true).
		Preload("Advertisements.File").
		First(&building, buildingId).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}

	if building.Advertisements == nil {
		building.Advertisements = []base_models.Advertisement{}
	}

	// 确保每个 Advertisement 的 File 都被正确加载
	for i := range building.Advertisements {
		if building.Advertisements[i].File.ID == 0 && building.Advertisements[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Advertisements[i].FileID).Error; err == nil {
				building.Advertisements[i].File = file
			}
		}
	}

	return building.Advertisements, nil
}

func (s *BuildingService) GetBuildingNotices(buildingId uint) ([]base_models.Notice, error) {
	var building base_models.Building
	if err := s.db.
		Preload("Notices", "is_public = ?", true).
		Preload("Notices.File").
		First(&building, buildingId).Error; err != nil {
		return nil, fmt.Errorf("building not found: %v", err)
	}

	if building.Notices == nil {
		building.Notices = []base_models.Notice{}
	}

	// 确保每个 Notice 的 File 都被正确加载
	for i := range building.Notices {
		if building.Notices[i].File.ID == 0 && building.Notices[i].FileID != nil {
			var file base_models.File
			if err := s.db.First(&file, building.Notices[i].FileID).Error; err == nil {
				building.Notices[i].File = file
			}
		}
		if building.Notices[i].FileType == "" {
			building.Notices[i].FileType = field.FileTypePdf
		}
	}

	return building.Notices, nil
}
