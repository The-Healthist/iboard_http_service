package building_admin_controllers

import (
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/gin-gonic/gin"
)

type BuildingAdminNoticeController struct {
	ctx           *gin.Context
	service       building_admin_services.InterfaceBuildingAdminNoticeService
	uploadService base_services.IUploadService
	fileService   base_services.InterfaceFileService
}

func NewBuildingAdminNoticeController(
	ctx *gin.Context,
	service building_admin_services.InterfaceBuildingAdminNoticeService,
	uploadService base_services.IUploadService,
	fileService base_services.InterfaceFileService,
) *BuildingAdminNoticeController {
	return &BuildingAdminNoticeController{
		ctx:           ctx,
		service:       service,
		uploadService: uploadService,
		fileService:   fileService,
	}
}

func (c *BuildingAdminNoticeController) GetNotices() {
	email := c.ctx.GetString("email")

	query := make(map[string]interface{})
	if noticeType := c.ctx.Query("type"); noticeType != "" {
		query["type"] = noticeType
	}

	// Parse pagination parameters
	pageSize := 10
	pageNum := 1
	if size := c.ctx.Query("pageSize"); size != "" {
		if parsed, err := strconv.Atoi(size); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if num := c.ctx.Query("pageNum"); num != "" {
		if parsed, err := strconv.Atoi(num); err == nil && parsed > 0 {
			pageNum = parsed
		}
	}

	paginate := map[string]interface{}{
		"pageSize": pageSize,
		"pageNum":  pageNum,
		"desc":     true,
	}

	notices, pagination, err := c.service.Get(email, query, paginate)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": pagination,
	})
}

func (c *BuildingAdminNoticeController) GetNotice() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	notice, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": notice})
}

func (c *BuildingAdminNoticeController) CreateNotice() {
	email := c.ctx.GetString("email")

	var notice base_models.Notice
	if err := c.ctx.ShouldBindJSON(&notice); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 设置 IsPublic 为 false
	notice.IsPublic = false

	if err := c.service.Create(&notice, email); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Notice created successfully",
		"data":    notice,
	})
}

func (c *BuildingAdminNoticeController) UpdateNotice() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ctx.ShouldBindJSON(&updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 确保不能修改为公开通知
	if isPublic, ok := updates["isPublic"].(bool); ok && isPublic {
		c.ctx.JSON(403, gin.H{"error": "Building admin cannot create public notices"})
		return
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Notice updated successfully"})
}

func (c *BuildingAdminNoticeController) DeleteNotice() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Notice deleted successfully"})
}

func (c *BuildingAdminNoticeController) GetUploadParams() {
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&req); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Missing required parameters"})
		return
	}

	// Get upload policy
	policy, err := c.uploadService.GetUploadParams(req.FileName)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": policy})
}
