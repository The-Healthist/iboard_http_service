package building_admin_views

import (
	building_admin_controllers "github.com/The-Healthist/iboard_http_service/controller/building_admin"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func GetNotices(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminNoticeService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminNoticeController(ctx, service)
	controller.GetNotices()
}

func GetNotice(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminNoticeService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminNoticeController(ctx, service)
	controller.GetNotice()
}

func CreateNotice(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminNoticeService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminNoticeController(ctx, service)
	controller.CreateNotice()
}

func UpdateNotice(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminNoticeService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminNoticeController(ctx, service)
	controller.UpdateNotice()
}

func DeleteNotice(ctx *gin.Context) {
	buildingAdminService := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	service := building_admin_services.NewBuildingAdminNoticeService(
		databases.DB_CONN,
		buildingAdminService,
	)
	controller := building_admin_controllers.NewBuildingAdminNoticeController(ctx, service)
	controller.DeleteNotice()
}

func RegisterBuildingAdminNoticeView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTBuildingAdmin())
	{
		r.GET("/notices", GetNotices)
		r.GET("/notices/:id", GetNotice)
		r.POST("/notices", CreateNotice)
		r.PUT("/notices/:id", UpdateNotice)
		r.DELETE("/notices/:id", DeleteNotice)
	}
}
