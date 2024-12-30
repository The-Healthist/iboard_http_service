package building_admin_controllers

import (
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/The-Healthist/iboard_http_service/utils/response"
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
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// 处理分页参数
	pageSize, _ := strconv.Atoi(c.ctx.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.ctx.DefaultQuery("pageNum", "1"))
	desc := c.ctx.DefaultQuery("desc", "true") == "true"

	// 处理查询参数
	query := make(map[string]interface{})
	if noticeType := c.ctx.Query("type"); noticeType != "" {
		query["type"] = field.NoticeType(noticeType)
	}
	if status := c.ctx.Query("status"); status != "" {
		query["status"] = field.Status(status)
	}
	if fileID := c.ctx.Query("fileId"); fileID != "" {
		if id, err := strconv.ParseUint(fileID, 10, 64); err == nil {
			query["fileId"] = uint(id)
		}
	}
	if fileType := c.ctx.Query("fileType"); fileType != "" {
		query["fileType"] = field.FileType(fileType)
	}

	paginate := map[string]interface{}{
		"pageSize": pageSize,
		"pageNum":  pageNum,
		"desc":     desc,
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
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	notice, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": notice})
}

type CreateNoticeRequest struct {
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Type        field.NoticeType `json:"type" binding:"required"`
	Status      field.Status     `json:"status" binding:"required"`
	StartTime   *time.Time       `json:"startTime" binding:"required"`
	EndTime     *time.Time       `json:"endTime" binding:"required"`
	IsPublic    bool             `json:"isPublic"`
	FileID      *uint            `json:"fileId"`
	FileType    field.FileType   `json:"fileType"`
}

func (c *BuildingAdminNoticeController) CreateNotice() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateNoticeRequest
	if err := c.ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c.ctx, err)
		return
	}

	notice := &base_models.Notice{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      req.Status,
		StartTime:   *req.StartTime,
		EndTime:     *req.EndTime,
		FileID:      req.FileID,
		FileType:    req.FileType,
		IsPublic:    false, // 强制设置为 false
	}

	if err := c.service.Create(notice, email); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Notice created successfully",
		"data":    notice,
	})
}

func (c *BuildingAdminNoticeController) UpdateNotice() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ctx.ShouldBindJSON(&updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 强制设置 isPublic 为 false
	if _, ok := updates["is_public"]; ok {
		updates["is_public"] = false
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Notice updated successfully"})
}

func (c *BuildingAdminNoticeController) DeleteNotice() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
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
