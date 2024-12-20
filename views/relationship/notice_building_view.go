package http_relationship_view

import (
	http_relationship_controller "github.com/The-Healthist/iboard_http_service/controller/relationship"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func BindBuildings_notice_building(ctx *gin.Context) {
	service := relationship_service.NewNoticeBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewNoticeBuildingController(ctx, service)
	controller.BindBuildings()
}

func UnbindBuildings_notice_building(ctx *gin.Context) {
	service := relationship_service.NewNoticeBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewNoticeBuildingController(ctx, service)
	controller.UnbindBuildings()
}

func GetBuildingsByNotice_notice_building(ctx *gin.Context) {
	service := relationship_service.NewNoticeBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewNoticeBuildingController(ctx, service)
	controller.GetBuildingsByNotice()
}

func GetNoticesByBuilding_notice_building(ctx *gin.Context) {
	service := relationship_service.NewNoticeBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewNoticeBuildingController(ctx, service)
	controller.GetNoticesByBuilding()
}

func RegisterNoticeBuildingView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/notice_building/bind", BindBuildings_notice_building)
		r.POST("/notice_building/unbind", UnbindBuildings_notice_building)
		r.GET("/notice_building/buildings", GetBuildingsByNotice_notice_building)
		r.GET("/notice_building/notices", GetNoticesByBuilding_notice_building)
	}
}
