package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		&jwtService,
		nil,
	)

	superAdminController.Login()
}

func CreateSuperAdmin(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.CreateSuperAdmin()
}

func GetSuperAdmins(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.GetSuperAdmins()
}

func DeleteSuperAdmin(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.DeleteSuperAdmin()
}

func ResetPassword(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	emailService := base_services.NewEmailService(utils.EmailClient)
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		&emailService,
	)

	superAdminController.ResetPassword()
}

func ChangePassword(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		&jwtService,
		nil,
	)

	superAdminController.ChangePassword()
}

func GetOneSuperAdmin(ctx *gin.Context) {
	superAdminService := base_services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_base_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.GetOne()
}

func RegisterSuperAdminView(r *gin.RouterGroup) {
	r.POST("/login", Login)
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/super_admin", CreateSuperAdmin)
		r.GET("/super_admin/:id", GetOneSuperAdmin)
		r.GET("/super_admin", GetSuperAdmins)
		r.DELETE("/super_admin", DeleteSuperAdmin)
		r.POST("/super_admin/reset_password", ResetPassword) //dont use
		r.POST("/super_admin/update_password", ChangePassword)
	}
}
