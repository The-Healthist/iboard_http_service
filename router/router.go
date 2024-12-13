package router

import (
	super_admin_views "github.com/The-Healthist/iboard_http_service/views"
	"github.com/gin-gonic/gin"
)

// Register Route
func RegisterRoute(r *gin.Engine) {
	// Register Admin Route
	super_admin_views.RegisterSuperAdminView(r.Group("/api/super_admin"))
}
