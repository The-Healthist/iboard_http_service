package http_relationship_view

import (
	http_relationship_controller "github.com/The-Healthist/iboard_http_service/controller/relationship"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

func BindFile_notice(ctx *gin.Context) {
	service := relationship_service.NewFileNoticeService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileNoticeController(ctx, service)
	controller.BindFile()
}

func UnbindFile_notice(ctx *gin.Context) {
	service := relationship_service.NewFileNoticeService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileNoticeController(ctx, service)
	controller.UnbindFile()
}

func GetNoticeByFile_notice(ctx *gin.Context) {
	service := relationship_service.NewFileNoticeService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileNoticeController(ctx, service)
	controller.GetNoticeByFile()
}

func GetFileByNotice_notice(ctx *gin.Context) {
	service := relationship_service.NewFileNoticeService(databases.DB_CONN)
	controller := http_relationship_controller.NewFileNoticeController(ctx, service)
	controller.GetFileByNotice()
}

func RegisterFileNoticeView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/file_notice/bind", BindFile_notice)
		r.POST("/file_notice/unbind", UnbindFile_notice)
		r.GET("/file_notice/notice", GetNoticeByFile_notice)
		r.GET("/file_notice/file", GetFileByNotice_notice)
	}
}
