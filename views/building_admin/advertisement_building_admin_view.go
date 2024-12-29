package building_admin_views

import (
	building_admin_controllers "github.com/The-Healthist/iboard_http_service/controller/building_admin"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func GetAdvertisements(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminAdvertisementService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminAdvertisementController(ctx, service)
	controller.GetAdvertisements()
}

func GetAdvertisement(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminAdvertisementService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminAdvertisementController(ctx, service)
	controller.GetAdvertisement()
}

func CreateAdvertisement(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminAdvertisementService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminAdvertisementController(ctx, service)
	controller.CreateAdvertisement()
}

func UpdateAdvertisement(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminAdvertisementService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminAdvertisementController(ctx, service)
	controller.UpdateAdvertisement()
}

func DeleteAdvertisement(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminAdvertisementService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminAdvertisementController(ctx, service)
	controller.DeleteAdvertisement()
}

func RegisterBuildingAdminAdvertisementView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTBuildingAdmin())
	{
		r.GET("/advertisements", GetAdvertisements)
		r.GET("/advertisements/:id", GetAdvertisement)
		r.POST("/advertisements", CreateAdvertisement)
		r.PUT("/advertisements/:id", UpdateAdvertisement)
		r.DELETE("/advertisements/:id", DeleteAdvertisement)
	}
}
