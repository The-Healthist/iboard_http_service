package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

// 1,CreateBuildingAdmin
func CreateBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := base_services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_base_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Create()
}

// 2, GetBuildingAdmins
func GetBuildingAdmins(ctx *gin.Context) {
	buildingAdminService := base_services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_base_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Get()
}

// 3, UpdateBuildingAdmin
func UpdateBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := base_services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_base_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Update()
}

// 4, DeleteBuildingAdmin
func DeleteBuildingAdmin(ctx *gin.Context) {
	buildingAdminService := base_services.NewBuildingAdminService(databases.DB_CONN)
	buildingAdminController := http_base_controller.NewBuildingAdminController(
		ctx,
		buildingAdminService,
	)

	buildingAdminController.Delete()
}

// 5, RegisterBuildingAdminView
func RegisterBuildingAdminView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/building_admin", CreateBuildingAdmin)
		r.GET("/building_admin", GetBuildingAdmins)
		r.PUT("/building_admin", UpdateBuildingAdmin)
		r.DELETE("/building_admin", DeleteBuildingAdmin)
	}
}
