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
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
	)

	fileController.Create()
}

func GetFiles(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
	)

	fileController.Get()
}

func UpdateFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
	)

	fileController.Update()
}

func DeleteFile(ctx *gin.Context) {
	fileService := base_services.NewFileService(databases.DB_CONN)
	fileController := http_base_controller.NewFileController(
		ctx,
		fileService,
	)

	fileController.Delete()
}

func RegisterFileView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateFile)
		r.GET("/get", GetFiles)
		r.PUT("/update", UpdateFile)
		r.DELETE("/delete", DeleteFile)
	}
}
