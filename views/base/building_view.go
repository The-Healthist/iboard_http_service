package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func CreateBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Create()
}

func GetBuildings(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Get()
}

func UpdateBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Update()
}

func DeleteBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Delete()
}

func RegisterBuildingView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/building", CreateBuilding)
		r.GET("/building", GetBuildings)
		r.PUT("/building", UpdateBuilding)
		r.DELETE("/building", DeleteBuilding)
	}
}
