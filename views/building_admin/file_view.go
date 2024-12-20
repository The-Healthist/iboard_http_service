package building_admin_views

import (
	building_admin_controllers "github.com/The-Healthist/iboard_http_service/controller/building_admin"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/gin-gonic/gin"
)

func GetFiles(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.GetFiles()
}

func GetFile(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.GetFile()
}

func UploadFile(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.UploadFile()
}

func UpdateFile(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.UpdateFile()
}

func DeleteFile(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.DeleteFile()
}

func DownloadFile(ctx *gin.Context) {
	service := building_admin_services.NewBuildingAdminFileService(databases.DB_CONN)
	controller := building_admin_controllers.NewBuildingAdminFileController(ctx, service)
	controller.DownloadFile()
}

func RegisterBuildingAdminFileView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTBuildingAdmin())
	{
		r.GET("/files", GetFiles)
		r.GET("/files/:id", GetFile)
		r.POST("/files/upload", UploadFile)
		r.PUT("/files/:id", UpdateFile)
		r.DELETE("/files/:id", DeleteFile)
		r.GET("/files/:id/download", DownloadFile)
	}
}
