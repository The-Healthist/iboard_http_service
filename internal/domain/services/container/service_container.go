package container

import (
	"context"
	"fmt"
	"sync"
	"time"

	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/building_admin"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	redis "github.com/The-Healthist/iboard_http_service/internal/infrastructure/redis"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/gorm"
)

// ServiceContainer holds all services
type ServiceContainer struct {
	db *gorm.DB

	// Base Services
	advertisementService base_services.InterfaceAdvertisementService
	buildingService      base_services.InterfaceBuildingService
	buildingAdminService base_services.InterfaceBuildingAdminService
	noticeService        base_services.InterfaceNoticeService
	fileService          base_services.InterfaceFileService
	jwtService           base_services.IJWTService
	emailService         base_services.IEmailService
	superAdminService    base_services.InterfaceSuperAdminService
	uploadService        base_services.IUploadService
	deviceService        base_services.InterfaceDeviceService
	noticeSyncService    base_services.InterfaceNoticeSyncService
    appService           base_services.InterfaceAppService

	// Building Admin Services
	buildingAdminAdvertisementService building_admin_services.InterfaceBuildingAdminAdvertisementService
	buildingAdminNoticeService        building_admin_services.InterfaceBuildingAdminNoticeService
	buildingAdminFileService          building_admin_services.InterfaceBuildingAdminFileService

	// Relationship Services
	advertisementBuildingService relationship_service.InterfaceAdvertisementBuildingService
	noticeBuildingService        relationship_service.InterfaceNoticeBuildingService
	buildingAdminBuildingService relationship_service.InterfaceBuildingAdminBuildingService
	fileAdvertisementService     relationship_service.InterfaceFileAdvertisementService
	fileNoticeService            relationship_service.InterfaceFileNoticeService
	deviceBuildingService        relationship_service.InterfaceDeviceBuildingService

	mu sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer(db *gorm.DB) *ServiceContainer {
	log.Info("创建服务容器...")

	if db == nil {
		log.Fatal("无法创建服务容器: 数据库连接为空")
		panic("数据库连接为空")
	}

	// Verify Redis connection
	if redis.REDIS_CONN == nil {
		log.Info("Redis连接未初始化，尝试初始化...")
		if err := redis.InitRedis(); err != nil {
			log.Fatal("Redis初始化失败: %v", err)
			panic(fmt.Sprintf("Redis初始化失败: %v", err))
		}
	}

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redis.REDIS_CONN.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis连接测试失败: %v", err)
		panic(fmt.Sprintf("Redis连接测试失败: %v", err))
	}

	log.Debug("Redis连接测试成功")

	container := &ServiceContainer{
		db: db,
	}
	container.initializeServices()
	log.Info("服务容器创建成功")
	return container
}

// initializeServices initializes all services
func (c *ServiceContainer) initializeServices() {
	log.Info("初始化服务...")

	if c.db == nil {
		log.Fatal("无法初始化服务: 数据库连接为空")
		panic("数据库连接为空")
	}

	// Ensure Redis is initialized
	if redis.REDIS_CONN == nil {
		log.Info("Redis连接未初始化，尝试初始化...")
		if err := redis.InitRedis(); err != nil {
			log.Fatal("Redis初始化失败: %v", err)
			panic(fmt.Sprintf("Redis初始化失败: %v", err))
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize Base Services
	log.Debug("初始化基础服务...")
	c.jwtService = base_services.NewJWTService()
	c.fileService = base_services.NewFileService(c.db)
	c.buildingService = base_services.NewBuildingService(c.db)
	c.advertisementService = base_services.NewAdvertisementService(c.db)
	c.noticeService = base_services.NewNoticeService(c.db)
	c.buildingAdminService = base_services.NewBuildingAdminService(c.db)
	c.superAdminService = base_services.NewSuperAdminService(c.db)
	c.deviceService = base_services.NewDeviceService(c.db)

	// Use global Redis connection
	c.uploadService = base_services.NewUploadService(c.db, redis.REDIS_CONN)

	// Initialize Notice Sync Service
	log.Debug("初始化通知同步服务...")
	c.noticeSyncService = base_services.NewNoticeSyncService(
		c.db,
		redis.REDIS_CONN,
		c.buildingService,
		c.uploadService,
		c.fileService,
	)
    // App service
    c.appService = base_services.NewAppService(c.db)

	// Initialize Relationship Services
	log.Debug("初始化关系服务...")
	c.buildingAdminBuildingService = relationship_service.NewBuildingAdminBuildingService(c.db)
	c.advertisementBuildingService = relationship_service.NewAdvertisementBuildingService(c.db)
	c.noticeBuildingService = relationship_service.NewNoticeBuildingService(c.db)
	c.fileAdvertisementService = relationship_service.NewFileAdvertisementService(c.db)
	c.fileNoticeService = relationship_service.NewFileNoticeService(c.db)
	c.deviceBuildingService = relationship_service.NewDeviceBuildingService(c.db)

	// Initialize Building Admin Services
	log.Debug("初始化建筑管理员服务...")
	c.buildingAdminAdvertisementService = building_admin_services.NewBuildingAdminAdvertisementService(
		c.db,
		c.buildingAdminBuildingService,
		c.fileService,
	)
	c.buildingAdminNoticeService = building_admin_services.NewBuildingAdminNoticeService(
		c.db,
		c.buildingAdminBuildingService,
		c.fileService,
	)
	c.buildingAdminFileService = building_admin_services.NewBuildingAdminFileService(c.db)

	log.Info("所有服务初始化完成")
}

// GetService returns the requested service based on the service name
func (c *ServiceContainer) GetService(name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	log.Debug("获取服务: %s", name)

	var service interface{}

	switch name {
	// Base services
	case "advertisement":
		service = c.advertisementService
	case "building":
		service = c.buildingService
	case "buildingAdmin":
		service = c.buildingAdminService
	case "notice":
		service = c.noticeService
	case "file":
		service = c.fileService
	case "jwt":
		service = c.jwtService
	case "superAdmin":
		service = c.superAdminService
	case "email":
		service = c.emailService
	case "upload":
		service = c.uploadService
	case "device":
		service = c.deviceService
	case "noticeSync":
		service = c.noticeSyncService
    case "app":
        service = c.appService

	// Building admin services
	case "buildingAdminAdvertisement":
		service = c.buildingAdminAdvertisementService
	case "buildingAdminNotice":
		service = c.buildingAdminNoticeService
	case "buildingAdminFile":
		service = c.buildingAdminFileService

	// Relationship services
	case "advertisementBuilding":
		service = c.advertisementBuildingService
	case "noticeBuilding":
		service = c.noticeBuildingService
	case "buildingAdminBuilding":
		service = c.buildingAdminBuildingService
	case "fileAdvertisement":
		service = c.fileAdvertisementService
	case "fileNotice":
		service = c.fileNoticeService
	case "deviceBuilding":
		service = c.deviceBuildingService
	default:
		log.Warn("请求的服务不存在: %s", name)
		return nil
	}

	if service == nil {
		log.Warn("服务未初始化: %s", name)
	}

	return service
}
