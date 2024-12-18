package views

import (
	http_controller "github.com/The-Healthist/iboard_http_service/controller"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/gin-gonic/gin"
)

// 1,CreateBuildingAdmin
func CreateBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Create()
}

// 2, GetBuildingAdmins
func GetBuildingAdmins(ctx *gin.Context) {
	buildingAdminService := services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Get()
}

// 3, UpdateBuildingAdmin
func UpdateBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Update()
}

// 4, DeleteBuildingAdmin
func DeleteBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Delete()
}

// 5, RegisterBuildingAdminView
func RegisterBuildingAdminView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateBuildingAdmin)
		r.GET("/get", GetBuildingAdmins)
		r.PUT("/update", UpdateBuildingAdmin)
		r.DELETE("/delete", DeleteBuildingAdmin)
	}
}
