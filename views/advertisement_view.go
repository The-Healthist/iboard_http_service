package views

import (
	http_controller "github.com/The-Healthist/iboard_http_service/controller"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/gin-gonic/gin"
)

func CreateAdvertisement(ctx *gin.Context) {
	advertisementService := services.NewAdvertisementService(databases.DB_CONN)
	advertisementController := http_controller.NewAdvertisementController(
		ctx,
		advertisementService,
	)

	advertisementController.Create()
}

func GetAdvertisements(ctx *gin.Context) {
	advertisementService := services.NewAdvertisementService(databases.DB_CONN)
	advertisementController := http_controller.NewAdvertisementController(
		ctx,
		advertisementService,
	)

	advertisementController.Get()
}

func UpdateAdvertisement(ctx *gin.Context) {
	advertisementService := services.NewAdvertisementService(databases.DB_CONN)
	advertisementController := http_controller.NewAdvertisementController(
		ctx,
		advertisementService,
	)

	advertisementController.Update()
}

func DeleteAdvertisement(ctx *gin.Context) {
	advertisementService := services.NewAdvertisementService(databases.DB_CONN)
	advertisementController := http_controller.NewAdvertisementController(
		ctx,
		advertisementService,
	)

	advertisementController.Delete()
}

func RegisterAdvertisementView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateAdvertisement)
		r.GET("/get", GetAdvertisements)
		r.PUT("/update", UpdateAdvertisement)
		r.DELETE("/delete", DeleteAdvertisement)
	}
}
