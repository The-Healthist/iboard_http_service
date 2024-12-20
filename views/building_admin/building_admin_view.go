package building_admin_views

import (
	building_admin_controllers "github.com/The-Healthist/iboard_http_service/controller/building_admin"
	databases "github.com/The-Healthist/iboard_http_service/database"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func BuildingAdminLogin(ctx *gin.Context) {
	buildingAdminService := base_services.NewBuildingAdminService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	controller := building_admin_controllers.NewBuildingAdminAuthController(
		ctx,
		buildingAdminService,
		jwtService,
	)
	controller.Login()
}

func RegisterBuildingAdminAuthView(r *gin.RouterGroup) {
	r.POST("/login", BuildingAdminLogin)
}
