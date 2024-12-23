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
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.Create()
}

func GetBuildings(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.Get()
}

func GetOneBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.GetOne()
}

func UpdateBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.Update()
}

func DeleteBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.Delete()
}

func LoginBuilding(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.Login()
}

func GetBuildingAdvertisements(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.GetBuildingAdvertisements()
}

func GetBuildingNotices(ctx *gin.Context) {
	buildingService := base_services.NewBuildingService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	buildingController := http_base_controller.NewBuildingController(
		ctx,
		buildingService,
		&jwtService,
	)

	buildingController.GetBuildingNotices()
}

func RegisterBuildingView(r *gin.RouterGroup) {
	// Public routes (no authentication required)
	r.POST("/building/login", LoginBuilding)

	// Building client routes (requires building JWT)
	buildingClient := r.Group("/building/client")
	buildingClient.Use(middlewares.AuthorizeJWTBuilding())
	{
		buildingClient.GET("/advertisements", GetBuildingAdvertisements)
		buildingClient.GET("/notices", GetBuildingNotices)
	}

	// Admin routes (requires admin JWT)
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/building", CreateBuilding)
		r.GET("/building", GetBuildings)
		r.GET("/building/:id", GetOneBuilding)
		r.PUT("/building", UpdateBuilding)
		r.DELETE("/building", DeleteBuilding)
	}
}
