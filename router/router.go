package router

import (
	base_views "github.com/The-Healthist/iboard_http_service/views/base"
	"github.com/gin-gonic/gin"
)

// Register Route
func RegisterRoute(r *gin.Engine) {
	// Register Admin Route
	base_views.RegisterSuperAdminView(r.Group("/api/super_admin"))
	base_views.RegisterBuildingAdminView(r.Group("/api/building_admin"))
	base_views.RegisterBuildingView(r.Group("/api/building"))
	base_views.RegisterFileView(r.Group("/api/file"))
	base_views.RegisterAdvertisementView(r.Group("/api/advertisement"))
	base_views.RegisterNoticeView(r.Group("/api/notice"))
}
