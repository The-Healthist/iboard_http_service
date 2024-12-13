package views

import (
	http_controller "github.com/The-Healthist/iboard_http_service/controller"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	jwtService := services.NewJWTService()
	superAdminController := http_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		&jwtService,
		nil,
	)

	superAdminController.Login()
}

func CreateSuperAdmin(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.CreateSuperAdmin()
}

func GetSuperAdmins(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.GetSuperAdmins()
}

func DeleteSuperAdmin(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	superAdminController := http_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		nil,
	)

	superAdminController.DeleteSuperAdmin()
}

func ResetPassword(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	emailService := services.NewEmailService(utils.EmailClient)
	superAdminController := http_controller.NewSuperAdminController(
		ctx,
		superAdminService,
		nil,
		&emailService,
	)

	superAdminController.ResetPassword()
}

func ChangePassword(ctx *gin.Context) {
	superAdminService := services.NewSuperAdminService(databases.DB_CONN)
	jwtService := services.NewJWTService()
	superAdminController := http_controller.NewSuperAdminController(
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
		r.POST("/create", CreateSuperAdmin)
		r.GET("/get", GetSuperAdmins)
		r.DELETE("/delete", DeleteSuperAdmin)
		r.POST("/reset_password", ResetPassword) //dont use
		r.POST("/update_password", ChangePassword)
	}
}
