package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func CreateFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.Create()
}

func GetFiles(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.Get()
}

func UpdateFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.Update()
}

func DeleteFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.Delete()
}

func GetOneFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.GetOne()
}

func RegisterFileView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/file", CreateFile)
		r.POST("/files", CreateManyFiles)
		r.GET("/file", GetFiles)
		r.GET("/file/:id", GetOneFile)
		r.PUT("/file", UpdateFile)
		r.DELETE("/file", DeleteFile)
	}
}

func CreateManyFiles(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	jwtService := base_services.NewJWTService()
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
		&jwtService,
	)

	fileController.CreateMany()
}
