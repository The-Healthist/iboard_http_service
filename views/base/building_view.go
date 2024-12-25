package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

// 1,create
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

// 2,get
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

// 3,get one
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

// 4,update
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

// 5,delete
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

// 6,login
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

// 7,get building advertisements
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

// 8,get building notices
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

// 9,register
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
