package base_services

import (
	"errors"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"gorm.io/gorm"
)

// InterfaceVersionService 版本服务接口
type InterfaceVersionService interface {
	// 创建版本
	Create(version *models.Version) (*models.Version, error)
	// 获取版本列表
	GetList(page, pageSize int) ([]*models.Version, int64, error)
	// 根据ID获取版本
	GetByID(id uint) (*models.Version, error)
	// 根据版本号获取版本
	GetByVersionNumber(versionNumber string) (*models.Version, error)
	// 更新版本
	Update(version *models.Version) (*models.Version, error)
	// 删除版本
	Delete(id uint) error
	// 获取活跃版本列表
	GetActiveVersions() ([]*models.Version, error)
}

// VersionService 版本服务实现
type VersionService struct {
	db *gorm.DB
}

// NewVersionService 创建版本服务
func NewVersionService(db *gorm.DB) InterfaceVersionService {
	return &VersionService{db: db}
}

// Create 创建版本
func (s *VersionService) Create(version *models.Version) (*models.Version, error) {
	if version == nil {
		return nil, errors.New("version cannot be nil")
	}

	// 检查版本号是否已存在
	var existingVersion models.Version
	if err := s.db.Where("version_number = ?", version.VersionNumber).First(&existingVersion).Error; err == nil {
		return nil, errors.New("version number already exists")
	}

	if err := s.db.Create(version).Error; err != nil {
		return nil, err
	}

	return version, nil
}

// GetList 获取版本列表
func (s *VersionService) GetList(page, pageSize int) ([]*models.Version, int64, error) {
	var versions []*models.Version
	var total int64

	// 获取总数
	if err := s.db.Model(&models.Version{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := s.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, err
	}

	return versions, total, nil
}

// GetByID 根据ID获取版本
func (s *VersionService) GetByID(id uint) (*models.Version, error) {
	var version models.Version
	if err := s.db.First(&version, id).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

// GetByVersionNumber 根据版本号获取版本
func (s *VersionService) GetByVersionNumber(versionNumber string) (*models.Version, error) {
	var version models.Version
	if err := s.db.Where("version_number = ?", versionNumber).First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

// Update 更新版本
func (s *VersionService) Update(version *models.Version) (*models.Version, error) {
	if version == nil {
		return nil, errors.New("version cannot be nil")
	}
	if version.ID == 0 {
		return nil, errors.New("version ID cannot be zero")
	}

	// 检查版本号是否与其他版本冲突
	var existingVersion models.Version
	if err := s.db.Where("version_number = ? AND id != ?", version.VersionNumber, version.ID).First(&existingVersion).Error; err == nil {
		return nil, errors.New("version number already exists")
	}

	if err := s.db.Model(&models.Version{}).Where("id = ?", version.ID).Updates(version).Error; err != nil {
		return nil, err
	}

	return s.GetByID(version.ID)
}

// Delete 删除版本
func (s *VersionService) Delete(id uint) error {
	// 检查是否被App表引用
	var app models.App
	if err := s.db.Where("current_version_id = ?", id).First(&app).Error; err == nil {
		return errors.New("cannot delete version that is currently in use")
	}

	return s.db.Delete(&models.Version{}, id).Error
}

// GetActiveVersions 获取活跃版本列表
func (s *VersionService) GetActiveVersions() ([]*models.Version, error) {
	var versions []*models.Version
	if err := s.db.Where("status = ?", "active").Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}
