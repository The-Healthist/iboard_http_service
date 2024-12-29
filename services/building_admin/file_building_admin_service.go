package building_admin_services

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceBuildingAdminFileService interface {
	Create(file *base_models.File, email string) error
	Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error)
	Update(id uint, email string, updates map[string]interface{}) error
	Delete(id uint, email string) error
	GetByID(id uint, email string) (*base_models.File, error)
}

type BuildingAdminFileService struct {
	db *gorm.DB
}

func NewBuildingAdminFileService(db *gorm.DB) InterfaceBuildingAdminFileService {
	return &BuildingAdminFileService{db: db}
}

func (s *BuildingAdminFileService) Create(file *base_models.File, email string) error {
	file.Uploader = email
	file.UploaderType = "building_admin"
	return s.db.Create(file).Error
}

func (s *BuildingAdminFileService) Get(email string, query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error) {
	var files []base_models.File
	var total int64
	db := s.db.Model(&base_models.File{}).Where("uploader = ? AND uploader_type = ?", email, "building_admin")

	if mimeType, ok := query["mimeType"].(string); ok && mimeType != "" {
		db = db.Where("mime_type = ?", mimeType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if err := db.Limit(pageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return files, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *BuildingAdminFileService) Update(id uint, email string, updates map[string]interface{}) error {
	result := s.db.Model(&base_models.File{}).
		Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) Delete(id uint, email string) error {
	result := s.db.Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		Delete(&base_models.File{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found or no permission")
	}
	return nil
}

func (s *BuildingAdminFileService) GetByID(id uint, email string) (*base_models.File, error) {
	var file base_models.File
	if err := s.db.Where("id = ? AND uploader = ? AND uploader_type = ?", id, email, "building_admin").
		First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
