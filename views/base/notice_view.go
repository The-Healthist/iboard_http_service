package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func CreateNotice(ctx *gin.Context) {
	noticeService := base_services.NewNoticeService(databases.DB_CONN)
	noticeController := http_base_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Create()
}

func GetNotices(ctx *gin.Context) {
	noticeService := base_services.NewNoticeService(databases.DB_CONN)
	noticeController := http_base_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Get()
}

func UpdateNotice(ctx *gin.Context) {
	noticeService := base_services.NewNoticeService(databases.DB_CONN)
	noticeController := http_base_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Update()
}

func DeleteNotice(ctx *gin.Context) {
	noticeService := base_services.NewNoticeService(databases.DB_CONN)
	noticeController := http_base_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Delete()
}

func RegisterNoticeView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/notice", CreateNotice)
		r.GET("/notice", GetNotices)
		r.PUT("/notice", UpdateNotice)
		r.DELETE("/notice", DeleteNotice)
	}
}
