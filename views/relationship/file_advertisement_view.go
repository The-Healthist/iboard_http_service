package http_relationship_view

import (
	http_relationship_controller "github.com/The-Healthist/iboard_http_service/controller/relationship"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func BindFile_advertisement(ctx *gin.Context) {
	service := relationship_service.NewFileAdvertisementService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileAdvertisementController(ctx, service)
	controller.BindFile()
}

func UnbindFile_advertisement(ctx *gin.Context) {
	service := relationship_service.NewFileAdvertisementService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileAdvertisementController(ctx, service)
	controller.UnbindFile()
}

func GetAdvertisementByFile_advertisement(ctx *gin.Context) {
	service := relationship_service.NewFileAdvertisementService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileAdvertisementController(ctx, service)
	controller.GetAdvertisementByFile()
}

func GetFileByAdvertisement_advertisement(ctx *gin.Context) {
	service := relationship_service.NewFileAdvertisementService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileAdvertisementController(ctx, service)
	controller.GetFileByAdvertisement()
}

func RegisterFileAdvertisementView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/file_advertisement/bind", BindFile_advertisement)
		r.POST("/file_advertisement/unbind", UnbindFile_advertisement)
		r.GET("/file_advertisement/advertisement", GetAdvertisementByFile_advertisement)
		r.GET("/file_advertisement/file", GetFileByAdvertisement_advertisement)
	}
}
