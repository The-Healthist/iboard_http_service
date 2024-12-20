package http_relationship_view

import (
	relationship_controller "github.com/The-Healthist/iboard_http_service/controller/relationship"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func BindBuildings(ctx *gin.Context) {
	service := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	controller := relationship_controller.NewBuildingAdminBuildingController(ctx, service)
	controller.BindBuildings()
}

func UnbindBuildings(ctx *gin.Context) {
	service := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	controller := relationship_controller.NewBuildingAdminBuildingController(ctx, service)
	controller.UnbindBuildings()
}

func GetBuildingsByAdmin(ctx *gin.Context) {
	service := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	controller := relationship_controller.NewBuildingAdminBuildingController(ctx, service)
	controller.GetBuildingsByAdmin()
}

func GetAdminsByBuilding(ctx *gin.Context) {
	service := relationship_service.NewBuildingAdminBuildingService(databases.DB_CONN)
	controller := relationship_controller.NewBuildingAdminBuildingController(ctx, service)
	controller.GetAdminsByBuilding()
}

func RegisterBuildingAdminBuildingView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/building_admin_building/bind", BindBuildings)
		r.POST("/building_admin_building/unbind", UnbindBuildings)
		r.GET("/building_admin_building/buildings", GetBuildingsByAdmin)
		r.GET("/building_admin_building/building_admin", GetAdminsByBuilding)
	}
}
