package router

import (
	_ "github.com/The-Healthist/iboard_http_service/docs/docs_swagger" // 导入 swagger 文档
	http_base_controller "github.com/The-Healthist/iboard_http_service/internal/app/controller/base"
	http_building_admin_controller "github.com/The-Healthist/iboard_http_service/internal/app/controller/building_admin"
	http_relationship_controller "github.com/The-Healthist/iboard_http_service/internal/app/controller/relationship"
	middlewares "github.com/The-Healthist/iboard_http_service/internal/app/middleware"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	databases "github.com/The-Healthist/iboard_http_service/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var serviceContainer *container.ServiceContainer

// Register Route
func RegisterRoute(r *gin.Engine) *gin.Engine {
	r.Use(middlewares.Cors())

	// Initialize service container after database connection is established
	serviceContainer = container.NewServiceContainer(databases.DB_CONN)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	r.POST("/api/device/login", http_base_controller.HandleFuncDevice(serviceContainer, "login"))
	r.POST("/api/building_admin/login", http_building_admin_controller.HandleFuncBuildingAdminAuth(serviceContainer, "login"))
	// Admin login
	r.POST("/api/admin/login", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "login"))
	r.GET("/api/app/version", http_base_controller.HandleFuncApp(serviceContainer, "get"))

	// Upload routes - 移除JWT认证以支持OSS回调
	r.POST("/api/admin/upload/params", http_base_controller.HandleFuncUpload(serviceContainer, "getUploadParams"))
	r.POST("/api/admin/upload/callback", http_base_controller.HandleFuncUpload(serviceContainer, "uploadCallback"))
	r.POST("/api/admin/upload/callback_sync", http_base_controller.HandleFuncUpload(serviceContainer, "uploadCallbackSync"))
	r.POST("/api/admin/upload/params_sync", http_base_controller.HandleFuncUpload(serviceContainer, "getUploadParamsSync"))

	// Admin routes
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middlewares.AuthorizeJWTAdmin())
	{
		// File routes
		adminGroup.POST("/file", http_base_controller.HandleFuncFile(serviceContainer, "create"))
		adminGroup.POST("/files", http_base_controller.HandleFuncFile(serviceContainer, "createMany"))
		adminGroup.GET("/file", http_base_controller.HandleFuncFile(serviceContainer, "get"))
		adminGroup.GET("/file/:id", http_base_controller.HandleFuncFile(serviceContainer, "getOne"))
		adminGroup.PUT("/file", http_base_controller.HandleFuncFile(serviceContainer, "update"))
		adminGroup.DELETE("/file", http_base_controller.HandleFuncFile(serviceContainer, "delete"))

		// Building Admin routes
		adminGroup.POST("/building_admin", http_base_controller.HandleFuncBuildingAdmin(serviceContainer, "create"))
		adminGroup.GET("/building_admin", http_base_controller.HandleFuncBuildingAdmin(serviceContainer, "get"))
		adminGroup.GET("/building_admin/:id", http_base_controller.HandleFuncBuildingAdmin(serviceContainer, "getOne"))
		adminGroup.PUT("/building_admin", http_base_controller.HandleFuncBuildingAdmin(serviceContainer, "update"))
		adminGroup.DELETE("/building_admin", http_base_controller.HandleFuncBuildingAdmin(serviceContainer, "delete"))

		// Super Admin routes
		adminGroup.POST("/super_admin", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "createSuperAdmin"))
		adminGroup.GET("/super_admin", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "getSuperAdmins"))
		adminGroup.GET("/super_admin/:id", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "getOne"))
		adminGroup.DELETE("/super_admin", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "deleteSuperAdmin"))
		adminGroup.POST("/super_admin/reset_password", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "resetPassword"))
		adminGroup.POST("/super_admin/update_password", http_base_controller.HandleFuncSuperAdmin(serviceContainer, "changePassword"))

		// Advertisement routes
		adminGroup.POST("/advertisement", http_base_controller.HandleFuncAdvertisement(serviceContainer, "create"))
		adminGroup.POST("/advertisements", http_base_controller.HandleFuncAdvertisement(serviceContainer, "createMany"))
		adminGroup.GET("/advertisement", http_base_controller.HandleFuncAdvertisement(serviceContainer, "get"))
		adminGroup.GET("/advertisement/:id", http_base_controller.HandleFuncAdvertisement(serviceContainer, "getOne"))
		adminGroup.PUT("/advertisement", http_base_controller.HandleFuncAdvertisement(serviceContainer, "update"))
		adminGroup.DELETE("/advertisement", http_base_controller.HandleFuncAdvertisement(serviceContainer, "delete"))

		// Notice routes
		adminGroup.POST("/notice", http_base_controller.HandleFuncNotice(serviceContainer, "create"))
		adminGroup.POST("/notices", http_base_controller.HandleFuncNotice(serviceContainer, "createMany"))
		adminGroup.GET("/notice", http_base_controller.HandleFuncNotice(serviceContainer, "get"))
		adminGroup.GET("/notice/:id", http_base_controller.HandleFuncNotice(serviceContainer, "getOne"))
		adminGroup.PUT("/notice", http_base_controller.HandleFuncNotice(serviceContainer, "update"))
		adminGroup.DELETE("/notice", http_base_controller.HandleFuncNotice(serviceContainer, "delete"))

		// Building routes
		adminGroup.POST("/building", http_base_controller.HandleFuncBuilding(serviceContainer, "create"))
		adminGroup.GET("/building", http_base_controller.HandleFuncBuilding(serviceContainer, "get"))
		adminGroup.GET("/building/:id", http_base_controller.HandleFuncBuilding(serviceContainer, "getOne"))
		adminGroup.PUT("/building", http_base_controller.HandleFuncBuilding(serviceContainer, "update"))
		adminGroup.DELETE("/building", http_base_controller.HandleFuncBuilding(serviceContainer, "delete"))
		adminGroup.POST("/building/:id/sync_notice", http_base_controller.HandleFuncBuilding(serviceContainer, "manualSyncNotice"))

		// Version routes
		adminGroup.POST("/version", http_base_controller.HandleFuncVersion(serviceContainer, "create"))
		adminGroup.GET("/versions", http_base_controller.HandleFuncVersion(serviceContainer, "getList"))
		adminGroup.GET("/version/:id", http_base_controller.HandleFuncVersion(serviceContainer, "getOne"))
		adminGroup.PUT("/version", http_base_controller.HandleFuncVersion(serviceContainer, "update"))
		adminGroup.DELETE("/version/:id", http_base_controller.HandleFuncVersion(serviceContainer, "delete"))
		adminGroup.GET("/versions/active", http_base_controller.HandleFuncVersion(serviceContainer, "getActive"))

		// App routes
		adminGroup.PUT("/app/version", http_base_controller.HandleFuncApp(serviceContainer, "update"))

		// Relationship routes
		// Building Admin Building routes
		adminGroup.POST("/building_admin_building/bind", http_relationship_controller.HandleFuncBuildingAdminBuilding(serviceContainer, "bindBuildings"))
		adminGroup.POST("/building_admin_building/unbind", http_relationship_controller.HandleFuncBuildingAdminBuilding(serviceContainer, "unbindBuildings"))
		adminGroup.GET("/building_admin_building/buildings", http_relationship_controller.HandleFuncBuildingAdminBuilding(serviceContainer, "getBuildingsByBuildingAdmin"))
		adminGroup.GET("/building_admin_building/admins", http_relationship_controller.HandleFuncBuildingAdminBuilding(serviceContainer, "getBuildingAdminsByBuilding"))

		// Advertisement Building routes
		adminGroup.POST("/advertisement_building/bind", http_relationship_controller.HandleFuncAdvertisementBuilding(serviceContainer, "bindBuildings"))
		adminGroup.POST("/advertisement_building/unbind", http_relationship_controller.HandleFuncAdvertisementBuilding(serviceContainer, "unbindBuildings"))
		adminGroup.GET("/advertisement_building/buildings", http_relationship_controller.HandleFuncAdvertisementBuilding(serviceContainer, "getBuildingsByAdvertisement"))
		adminGroup.GET("/advertisement_building/advertisements", http_relationship_controller.HandleFuncAdvertisementBuilding(serviceContainer, "getAdvertisementsByBuilding"))

		// Notice Building routes
		adminGroup.POST("/notice_building/bind", http_relationship_controller.HandleFuncNoticeBuilding(serviceContainer, "bindBuildings"))
		adminGroup.POST("/notice_building/unbind", http_relationship_controller.HandleFuncNoticeBuilding(serviceContainer, "unbindBuildings"))
		adminGroup.GET("/notice_building/buildings", http_relationship_controller.HandleFuncNoticeBuilding(serviceContainer, "getBuildingsByNotice"))
		adminGroup.GET("/notice_building/notices", http_relationship_controller.HandleFuncNoticeBuilding(serviceContainer, "getNoticesByBuilding"))

		// File Notice routes
		adminGroup.POST("/file_notice/bind", http_relationship_controller.HandleFuncFileNotice(serviceContainer, "bindFile"))
		adminGroup.POST("/file_notice/unbind", http_relationship_controller.HandleFuncFileNotice(serviceContainer, "unbindFile"))
		adminGroup.GET("/file_notice/notice", http_relationship_controller.HandleFuncFileNotice(serviceContainer, "getNoticeByFile"))
		adminGroup.GET("/file_notice/file", http_relationship_controller.HandleFuncFileNotice(serviceContainer, "getFileByNotice"))

		// File Advertisement routes
		adminGroup.POST("/file_advertisement/bind", http_relationship_controller.HandleFuncFileAdvertisement(serviceContainer, "bindFile"))
		adminGroup.POST("/file_advertisement/unbind", http_relationship_controller.HandleFuncFileAdvertisement(serviceContainer, "unbindFile"))
		adminGroup.GET("/file_advertisement/advertisement", http_relationship_controller.HandleFuncFileAdvertisement(serviceContainer, "getAdvertisementByFile"))
		adminGroup.GET("/file_advertisement/file", http_relationship_controller.HandleFuncFileAdvertisement(serviceContainer, "getFileByAdvertisement"))

		// Device routes
		adminGroup.POST("/device", http_base_controller.HandleFuncDevice(serviceContainer, "create"))
		adminGroup.POST("/devices", http_base_controller.HandleFuncDevice(serviceContainer, "createMany"))
		adminGroup.GET("/device", http_base_controller.HandleFuncDevice(serviceContainer, "get"))
		adminGroup.PUT("/device", http_base_controller.HandleFuncDevice(serviceContainer, "update"))
		adminGroup.DELETE("/device", http_base_controller.HandleFuncDevice(serviceContainer, "delete"))
		adminGroup.GET("/device/:id", http_base_controller.HandleFuncDevice(serviceContainer, "getOne"))

		// Device-Building relationship routes
		adminGroup.POST("/device_building/bind", http_relationship_controller.HandleFuncDeviceBuilding(serviceContainer, "bindDevice"))
		adminGroup.POST("/device_building/unbind", http_relationship_controller.HandleFuncDeviceBuilding(serviceContainer, "unbindDevice"))
		adminGroup.GET("/device_building/devices", http_relationship_controller.HandleFuncDeviceBuilding(serviceContainer, "getDevicesByBuilding"))
		adminGroup.GET("/device_building/building", http_relationship_controller.HandleFuncDeviceBuilding(serviceContainer, "getBuildingByDevice"))

		//1.1.0 Admin set carousel orders (admin can view and update complete data)
		adminGroup.POST("/device/carousel/top_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getTopAdCarouselResolved"))
		adminGroup.PUT("/device/carousel/top_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "updateTopAdCarousel"))
		adminGroup.POST("/device/carousel/full_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getFullAdCarouselResolved"))
		adminGroup.PUT("/device/carousel/full_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "updateFullAdCarousel"))
		adminGroup.POST("/device/carousel/notices", http_base_controller.HandleFuncDevice(serviceContainer, "getNoticeCarouselResolved"))
		adminGroup.PUT("/device/carousel/notices", http_base_controller.HandleFuncDevice(serviceContainer, "updateNoticeCarousel"))
	}

	// Building admin routes (requires building admin JWT)
	buildingAdminGroup := r.Group("/api/building_admin")
	buildingAdminGroup.Use(middlewares.AuthorizeJWTBuildingAdmin())
	{
		// File routes
		buildingAdminGroup.GET("/file", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "getFiles"))
		buildingAdminGroup.GET("/file/:id", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "getFile"))
		buildingAdminGroup.POST("/file", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "uploadFile"))
		buildingAdminGroup.PUT("/file", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "updateFile"))
		buildingAdminGroup.DELETE("/file/:id", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "deleteFile"))
		buildingAdminGroup.GET("/file/:id/download", http_building_admin_controller.HandleFuncBuildingAdminFile(serviceContainer, "downloadFile"))

		// Advertisement routes
		buildingAdminGroup.GET("/advertisement", http_building_admin_controller.HandleFuncBuildingAdminAdvertisement(serviceContainer, "getAdvertisements"))
		buildingAdminGroup.GET("/advertisement/:id", http_building_admin_controller.HandleFuncBuildingAdminAdvertisement(serviceContainer, "getAdvertisement"))
		buildingAdminGroup.POST("/advertisement", http_building_admin_controller.HandleFuncBuildingAdminAdvertisement(serviceContainer, "createAdvertisement"))
		buildingAdminGroup.PUT("/advertisement", http_building_admin_controller.HandleFuncBuildingAdminAdvertisement(serviceContainer, "updateAdvertisement"))
		buildingAdminGroup.DELETE("/advertisement/:id", http_building_admin_controller.HandleFuncBuildingAdminAdvertisement(serviceContainer, "deleteAdvertisement"))

		// Notice routes
		buildingAdminGroup.GET("/notice", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "getNotices"))
		buildingAdminGroup.GET("/notice/:id", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "getNotice"))
		buildingAdminGroup.POST("/notice", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "createNotice"))
		buildingAdminGroup.PUT("/notice", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "updateNotice"))
		buildingAdminGroup.DELETE("/notice/:id", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "deleteNotice"))
		buildingAdminGroup.POST("/notice/upload/params", http_building_admin_controller.HandleFuncBuildingAdminNotice(serviceContainer, "getUploadParams"))
	}

	// Device client routes (requires device JWT)
	deviceClientGroup := r.Group("/api/device/client")
	deviceClientGroup.Use(middlewares.AuthorizeJWTDevice())
	{
		deviceClientGroup.GET("/advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getDeviceAdvertisements"))
		deviceClientGroup.GET("/notices", http_base_controller.HandleFuncDevice(serviceContainer, "getDeviceNotices"))
		deviceClientGroup.POST("/health_test", http_base_controller.HandleFuncDevice(serviceContainer, "healthTest"))
		//1.1.0
		deviceClientGroup.GET("/top_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getDeviceTopAdvertisements"))
		deviceClientGroup.GET("/full_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getDeviceFullAdvertisements"))

		deviceClientGroup.GET("/carousel/top_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getTopAdCarouselResolved"))
		deviceClientGroup.GET("/carousel/full_advertisements", http_base_controller.HandleFuncDevice(serviceContainer, "getFullAdCarouselResolved"))
		deviceClientGroup.GET("/carousel/notices", http_base_controller.HandleFuncDevice(serviceContainer, "getNoticeCarouselResolved"))
	}

	return r
}
