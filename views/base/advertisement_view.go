package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func CreateAdvertisement(ctx *gin.Context) {
	advertisementService := base_services.NewAdvertisementService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	advertisementController := http_base_controller.NewAdvertisementController(
		ctx,
		advertisementService,
		&jwtService,
	)

	advertisementController.Create()
}

func GetAdvertisements(ctx *gin.Context) {
	advertisementService := base_services.NewAdvertisementService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	advertisementController := http_base_controller.NewAdvertisementController(
		ctx,
		advertisementService,
		&jwtService,
	)

	advertisementController.Get()
}

func UpdateAdvertisement(ctx *gin.Context) {
	advertisementService := base_services.NewAdvertisementService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	advertisementController := http_base_controller.NewAdvertisementController(
		ctx,
		advertisementService,
		&jwtService,
	)

	advertisementController.Update()
}

func DeleteAdvertisement(ctx *gin.Context) {
	advertisementService := base_services.NewAdvertisementService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	advertisementController := http_base_controller.NewAdvertisementController(
		ctx,
		advertisementService,
		&jwtService,
	)

	advertisementController.Delete()
}

func GetOne(ctx *gin.Context) {
	advertisementService := base_services.NewAdvertisementService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	advertisementController := http_base_controller.NewAdvertisementController(
		ctx,
		advertisementService,
		&jwtService,
	)

	advertisementController.GetOne()
}

func RegisterAdvertisementView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/advertisement", CreateAdvertisement)
		r.GET("/advertisement", GetAdvertisements)
		r.GET("/advertisement/:id", GetOne)
		r.PUT("/advertisement", UpdateAdvertisement)
		r.DELETE("/advertisement", DeleteAdvertisement)
	}
}
