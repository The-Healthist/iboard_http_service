package base_views

import (
	http_base_controller "github.com/The-Healthist/iboard_http_service/controller/base"
	databases "github.com/The-Healthist/iboard_http_service/database"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

func GetUploadParams(ctx *gin.Context) {
	uploadService := base_services.NewUploadService(databases.DB_CONN)
	controller := http_base_controller.NewUploadController(ctx, uploadService)
	controller.GetUploadParams()
}

func RegisterUploadView(r *gin.RouterGroup) {
	r.POST("/upload/policy", GetUploadParams)
}
