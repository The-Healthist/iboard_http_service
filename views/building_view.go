package views

import (
	http_controller "github.com/The-Healthist/iboard_http_service/controller"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/gin-gonic/gin"
)

func CreateBuilding(ctx *gin.Context) {
	buildingService := services.NewBuildingService(databases.DB_CONN)
	buildingController := http_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Create()
}

func GetBuildings(ctx *gin.Context) {
	buildingService := services.NewBuildingService(databases.DB_CONN)
	buildingController := http_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Get()
}

func UpdateBuilding(ctx *gin.Context) {
	buildingService := services.NewBuildingService(databases.DB_CONN)
	buildingController := http_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Update()
}

func DeleteBuilding(ctx *gin.Context) {
	buildingService := services.NewBuildingService(databases.DB_CONN)
	buildingController := http_controller.NewBuildingController(
		ctx,
		buildingService,
	)

	buildingController.Delete()
}

func RegisterBuildingView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateBuilding)
		r.GET("/get", GetBuildings)
		r.PUT("/update", UpdateBuilding)
		r.DELETE("/delete", DeleteBuilding)
	}
}
