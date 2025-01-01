package container

import (
	"sync"

	"os"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/go-redis/redis/v8"
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

	mu sync.RWMutex
}

// NewServiceContainer creates a new service container
func NewServiceContainer(db *gorm.DB) *ServiceContainer {
	if db == nil {
		panic("database connection is nil")
	}
	container := &ServiceContainer{
		db: db,
	}
	container.initializeServices()
	return container
}

// initializeServices initializes all services
func (c *ServiceContainer) initializeServices() {
	if c.db == nil {
		panic("database connection is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize Base Services
	c.jwtService = base_services.NewJWTService()
	c.fileService = base_services.NewFileService(c.db)
	c.buildingService = base_services.NewBuildingService(c.db)
	c.advertisementService = base_services.NewAdvertisementService(c.db)
	c.noticeService = base_services.NewNoticeService(c.db)
	c.buildingAdminService = base_services.NewBuildingAdminService(c.db)
	c.superAdminService = base_services.NewSuperAdminService(c.db)
	c.uploadService = base_services.NewUploadService(c.db, c.getRedisClient())

	// Initialize Relationship Services
	c.buildingAdminBuildingService = relationship_service.NewBuildingAdminBuildingService(c.db)
	c.advertisementBuildingService = relationship_service.NewAdvertisementBuildingService(c.db)
	c.noticeBuildingService = relationship_service.NewNoticeBuildingService(c.db)
	c.fileAdvertisementService = relationship_service.NewFileAdvertisementService(c.db)
	c.fileNoticeService = relationship_service.NewFileNoticeService(c.db)

	// Initialize Building Admin Services
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
}

// GetService returns the requested service based on the service name
func (c *ServiceContainer) GetService(name string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	switch name {
	// Base services
	case "advertisement":
		return c.advertisementService
	case "building":
		return c.buildingService
	case "buildingAdmin":
		return c.buildingAdminService
	case "notice":
		return c.noticeService
	case "file":
		return c.fileService
	case "jwt":
		return c.jwtService
	case "superAdmin":
		return c.superAdminService
	case "email":
		return c.emailService
	case "upload":
		return c.uploadService

	// Building admin services
	case "buildingAdminAdvertisement":
		return c.buildingAdminAdvertisementService
	case "buildingAdminNotice":
		return c.buildingAdminNoticeService
	case "buildingAdminFile":
		return c.buildingAdminFileService

	// Relationship services
	case "advertisementBuilding":
		return c.advertisementBuildingService
	case "noticeBuilding":
		return c.noticeBuildingService
	case "buildingAdminBuilding":
		return c.buildingAdminBuildingService
	case "fileAdvertisement":
		return c.fileAdvertisementService
	case "fileNotice":
		return c.fileNoticeService
	default:
		return nil
	}
}

// For backward compatibility and type safety, we'll keep the typed getters
// but they'll use GetService internally
// func (c *ServiceContainer) GetAdvertisementService() base_services.InterfaceAdvertisementService {
// 	return c.GetService("advertisement").(base_services.InterfaceAdvertisementService)
// }

// func (c *ServiceContainer) GetBuildingService() base_services.InterfaceBuildingService {
// 	return c.GetService("building").(base_services.InterfaceBuildingService)
// }

// func (c *ServiceContainer) GetBuildingAdminService() base_services.InterfaceBuildingAdminService {
// 	return c.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService)
// }

// func (c *ServiceContainer) GetNoticeService() base_services.InterfaceNoticeService {
// 	return c.GetService("notice").(base_services.InterfaceNoticeService)
// }

// func (c *ServiceContainer) GetFileService() base_services.InterfaceFileService {
// 	return c.GetService("file").(base_services.InterfaceFileService)
// }

// func (c *ServiceContainer) GetJWTService() base_services.IJWTService {
// 	return c.GetService("jwt").(base_services.IJWTService)
// }

// func (c *ServiceContainer) GetSuperAdminService() base_services.InterfaceSuperAdminService {
// 	return c.GetService("superAdmin").(base_services.InterfaceSuperAdminService)
// }

// func (c *ServiceContainer) GetEmailService() base_services.IEmailService {
// 	return c.GetService("email").(base_services.IEmailService)
// }

// func (c *ServiceContainer) GetBuildingAdminAdvertisementService() building_admin_services.InterfaceBuildingAdminAdvertisementService {
// 	return c.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService)
// }

// func (c *ServiceContainer) GetBuildingAdminNoticeService() building_admin_services.InterfaceBuildingAdminNoticeService {
// 	return c.GetService("buildingAdminNotice").(building_admin_services.InterfaceBuildingAdminNoticeService)
// }

// func (c *ServiceContainer) GetBuildingAdminFileService() building_admin_services.InterfaceBuildingAdminFileService {
// 	return c.GetService("buildingAdminFile").(building_admin_services.InterfaceBuildingAdminFileService)
// }

// func (c *ServiceContainer) GetAdvertisementBuildingService() relationship_service.InterfaceAdvertisementBuildingService {
// 	return c.GetService("advertisementBuilding").(relationship_service.InterfaceAdvertisementBuildingService)
// }

// func (c *ServiceContainer) GetNoticeBuildingService() relationship_service.InterfaceNoticeBuildingService {
// 	return c.GetService("noticeBuilding").(relationship_service.InterfaceNoticeBuildingService)
// }

// func (c *ServiceContainer) GetBuildingAdminBuildingService() relationship_service.InterfaceBuildingAdminBuildingService {
// 	return c.GetService("buildingAdminBuilding").(relationship_service.InterfaceBuildingAdminBuildingService)
// }

// func (c *ServiceContainer) GetFileAdvertisementService() relationship_service.InterfaceFileAdvertisementService {
// 	return c.GetService("fileAdvertisement").(relationship_service.InterfaceFileAdvertisementService)
// }

// func (c *ServiceContainer) GetFileNoticeService() relationship_service.InterfaceFileNoticeService {
// 	return c.GetService("fileNotice").(relationship_service.InterfaceFileNoticeService)
// }

// Helper function to get Redis client
func (c *ServiceContainer) getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}
