package router

import (
	base_views "github.com/The-Healthist/iboard_http_service/views/base"
	building_admin_views "github.com/The-Healthist/iboard_http_service/views/building_admin"
	relationship_views "github.com/The-Healthist/iboard_http_service/views/relationship"
	"github.com/gin-gonic/gin"
)

// Register Route
func RegisterRoute(r *gin.Engine) {
	// Register Admin Route
	base_views.RegisterUploadView(r.Group("/api/admin"))
	base_views.RegisterSuperAdminView(r.Group("/api/admin"))
	base_views.RegisterBuildingAdminView(r.Group("/api/admin"))
	base_views.RegisterBuildingView(r.Group("/api/admin"))
	base_views.RegisterFileView(r.Group("/api/admin"))
	base_views.RegisterAdvertisementView(r.Group("/api/admin"))
	base_views.RegisterNoticeView(r.Group("/api/admin"))
	relationship_views.RegisterBuildingAdminBuildingView(r.Group("/api/admin"))
	relationship_views.RegisterAdvertisementBuildingView(r.Group("/api/admin"))
	relationship_views.RegisterNoticeBuildingView(r.Group("/api/admin"))
	relationship_views.RegisterFileNoticeView(r.Group("/api/admin"))
	relationship_views.RegisterFileAdvertisementView(r.Group("/api/admin"))

	// BuildingAdmin 路由组
	buildingAdminGroup := r.Group("/api/building_admin")
	{
		building_admin_views.RegisterBuildingAdminAuthView(buildingAdminGroup)
		building_admin_views.RegisterBuildingAdminFileView(buildingAdminGroup)
		building_admin_views.RegisterBuildingAdminAdvertisementView(buildingAdminGroup)
		building_admin_views.RegisterBuildingAdminNoticeView(buildingAdminGroup)
	}
}
