package base_services

import (
	"errors"
	"time"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"gorm.io/gorm"
)

// InterfaceAppService 应用服务接口
type InterfaceAppService interface {
	// 获取当前版本配置
	Get() (*models.App, error)
	// 更新版本配置
	Update(app *models.App) (*models.App, error)
}

// AppService 应用服务实现
type AppService struct {
	db *gorm.DB
}

// NewAppService 创建应用服务
func NewAppService(db *gorm.DB) InterfaceAppService {
	return &AppService{db: db}
}

// Get 获取当前版本配置
func (s *AppService) Get() (*models.App, error) {
	var app models.App
	if err := s.db.Preload("CurrentVersion").First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 初始化一条默认记录
			app = models.App{
				CurrentVersionID: 1, // 默认使用第一个版本
				LastCheckTime:    time.Now(),
				UpdateInterval:   3600,
				AutoUpdate:       false,
				Status:           "active",
			}
			if err := s.db.Create(&app).Error; err != nil {
				return nil, err
			}
			return &app, nil
		}
		return nil, err
	}
	return &app, nil
}

// Update 更新版本配置
func (s *AppService) Update(app *models.App) (*models.App, error) {
	if app == nil {
		return nil, errors.New("app cannot be nil")
	}
	if app.ID == 0 {
		return nil, errors.New("app ID cannot be zero")
	}

	// 如果设置了新的当前版本ID，验证版本是否存在
	if app.CurrentVersionID != 0 {
		var version models.Version
		if err := s.db.First(&version, app.CurrentVersionID).Error; err != nil {
			return nil, errors.New("version not found")
		}
	}

	// 更新应用信息
	if err := s.db.Model(&models.App{}).Where("id = ?", app.ID).Updates(app).Error; err != nil {
		return nil, err
	}

	// 返回更新后的应用信息
	return s.Get()
}
