package http_relationship_view

import (
	http_relationship_controller "github.com/The-Healthist/iboard_http_service/controller/relationship"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func BindBuildings_advertisement_building(ctx *gin.Context) {
	service := relationship_service.NewAdvertisementBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewAdvertisementBuildingController(ctx, service)
	controller.BindBuildings()
}

func UnbindBuildings_advertisement_building(ctx *gin.Context) {
	service := relationship_service.NewAdvertisementBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewAdvertisementBuildingController(ctx, service)
	controller.UnbindBuildings()
}

func GetBuildingsByAdvertisement_advertisement_building(ctx *gin.Context) {
	service := relationship_service.NewAdvertisementBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewAdvertisementBuildingController(ctx, service)
	controller.GetBuildingsByAdvertisement()
}

func GetAdvertisementsByBuilding_advertisement_building(ctx *gin.Context) {
	service := relationship_service.NewAdvertisementBuildingService(databases.DB_CONN)
	controller := http_relationship_controller.NewAdvertisementBuildingController(ctx, service)
	controller.GetAdvertisementsByBuilding()
}

func RegisterAdvertisementBuildingView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/advertisement_building/bind", BindBuildings_advertisement_building)
		r.POST("/advertisement_building/unbind", UnbindBuildings_advertisement_building)
		r.GET("/advertisement_building/buildings", GetBuildingsByAdvertisement_advertisement_building)
		r.GET("/advertisement_building/advertisements", GetAdvertisementsByBuilding_advertisement_building)
	}
}
