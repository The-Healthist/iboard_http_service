package views

import (
	http_controller "github.com/The-Healthist/iboard_http_service/controller"
	databases "github.com/The-Healthist/iboard_http_service/database"
	middlewares "github.com/The-Healthist/iboard_http_service/middleware"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/gin-gonic/gin"
)

func CreateNotice(ctx *gin.Context) {
	noticeService := services.NewNoticeService(databases.DB_CONN)
	noticeController := http_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Create()
}

func GetNotices(ctx *gin.Context) {
	noticeService := services.NewNoticeService(databases.DB_CONN)
	noticeController := http_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Get()
}

func UpdateNotice(ctx *gin.Context) {
	noticeService := services.NewNoticeService(databases.DB_CONN)
	noticeController := http_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Update()
}

func DeleteNotice(ctx *gin.Context) {
	noticeService := services.NewNoticeService(databases.DB_CONN)
	noticeController := http_controller.NewNoticeController(
		ctx,
		noticeService,
	)

	noticeController.Delete()
}

func RegisterNoticeView(r *gin.RouterGroup) {
	r.Use(middlewares.AuthorizeJWTAdmin())
	{
		r.POST("/add", CreateNotice)
		r.GET("/get", GetNotices)
		r.PUT("/update", UpdateNotice)
		r.DELETE("/delete", DeleteNotice)
	}
}
