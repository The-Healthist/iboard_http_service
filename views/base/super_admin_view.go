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

func RegisterSuperAdminView(r *gin.RouterGroup) {
	r.POST("/login", Login)
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateSuperAdmin)
		r.GET("/get", GetSuperAdmins)
		r.DELETE("/delete", DeleteSuperAdmin)
		r.POST("/reset_password", ResetPassword) //dont use
		r.POST("/update_password", ChangePassword)
	}
}
