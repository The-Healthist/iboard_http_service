package router

import (
	"github.com/The-Healthist/iboard_http_service/views"
	"github.com/gin-gonic/gin"
)

// Register Route
func RegisterRoute(r *gin.Engine) {
	// Register Admin Route
	views.RegisterSuperAdminView(r.Group("/api/super_admin"))
	views.RegisterBuildingAdminView(r.Group("/api/building_admin"))
	views.RegisterBuildingView(r.Group("/api/building"))
	views.RegisterFileView(r.Group("/api/file"))
	views.RegisterAdvertisementView(r.Group("/api/advertisement"))
}
