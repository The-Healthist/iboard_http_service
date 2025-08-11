package base_services

import (
    "errors"

    "github.com/The-Healthist/iboard_http_service/internal/domain/models"
    "gorm.io/gorm"
)

// 1.Interface 应用服务接口
type InterfaceAppService interface {
    // 1.Get 获取应用版本
    Get() (*models.App, error)
    // 2.Update 更新应用版本
    Update(version string) (*models.App, error)
}

// AppService 应用服务实现
type AppService struct {
    db *gorm.DB
}

// NewAppService 创建应用服务
func NewAppService(db *gorm.DB) InterfaceAppService {
    return &AppService{db: db}
}

// 1.Get 获取应用版本
func (s *AppService) Get() (*models.App, error) {
    var app models.App
    if err := s.db.First(&app).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 初始化一条默认记录
            app = models.App{Version: "1.0.0+0"}
            if err := s.db.Create(&app).Error; err != nil {
                return nil, err
            }
            return &app, nil
        }
        return nil, err
    }
    return &app, nil
}

// 2.Update 更新应用版本
func (s *AppService) Update(version string) (*models.App, error) {
    if version == "" {
        return nil, errors.New("version cannot be empty")
    }
    app, err := s.Get()
    if err != nil {
        return nil, err
    }
    if err := s.db.Model(app).Update("version", version).Error; err != nil {
        return nil, err
    }
    return app, nil
}


