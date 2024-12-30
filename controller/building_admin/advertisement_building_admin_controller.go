package building_admin_controllers

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/The-Healthist/iboard_http_service/utils/response"
	"github.com/gin-gonic/gin"
)

type BuildingAdminAdvertisementController struct {
	ctx     *gin.Context
	service building_admin_services.InterfaceBuildingAdminAdvertisementService
}

func NewBuildingAdminAdvertisementController(
	ctx *gin.Context,
	service building_admin_services.InterfaceBuildingAdminAdvertisementService,
) *BuildingAdminAdvertisementController {
	return &BuildingAdminAdvertisementController{
		ctx:     ctx,
		service: service,
	}
}

func (c *BuildingAdminAdvertisementController) GetAdvertisements() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var searchQuery struct {
		Search string `form:"search"`
		Type   string `form:"type"`
	}
	if err := c.ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pagination := struct {
		PageSize int  `form:"pageSize"`
		PageNum  int  `form:"pageNum"`
		Desc     bool `form:"desc"`
	}{
		PageSize: 10,
		PageNum:  1,
		Desc:     false,
	}

	if err := c.ctx.ShouldBindQuery(&pagination); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	queryMap := utils.StructToMap(searchQuery)
	paginationMap := map[string]interface{}{
		"pageSize": pagination.PageSize,
		"pageNum":  pagination.PageNum,
		"desc":     pagination.Desc,
	}

	advertisements, paginationResult, err := c.service.Get(email, queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       advertisements,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminAdvertisementController) GetAdvertisement() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	advertisement, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": advertisement})
}

type CreateAdvertisementRequest struct {
	Title       string                     `json:"title" binding:"required"`
	Description string                     `json:"description"`
	Type        field.AdvertisementType    `json:"type"`
	Status      field.Status               `json:"status"`
	Duration    int                        `json:"duration"`
	StartTime   *time.Time                 `json:"startTime"`
	EndTime     *time.Time                 `json:"endTime"`
	Display     field.AdvertisementDisplay `json:"display"`
	IsPublic    bool                       `json:"isPublic"`
	Path        string                     `json:"path"`
}

func (c *BuildingAdminAdvertisementController) CreateAdvertisement() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateAdvertisementRequest
	if err := c.ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c.ctx, err)
		return
	}

	// Set default values for startTime, endTime, and status
	now := time.Now()
	startTime := now
	endTime := now.AddDate(1, 0, 0) // 1 year from now
	status := field.Status("active")

	if req.StartTime != nil {
		startTime = *req.StartTime
	}
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	if req.Status != "" {
		status = req.Status
	}

	// 如果提供了 path，查找对应的文件
	var fileID *uint
	if req.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", req.Path).First(&file).Error; err != nil {
			c.ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID
	}

	advertisement := &base_models.Advertisement{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      status,
		Duration:    req.Duration,
		StartTime:   startTime,
		EndTime:     endTime,
		Display:     req.Display,
		FileID:      fileID,
		IsPublic:    false, // 强制设置为 false
	}

	if err := c.service.Create(advertisement, email); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create advertisement failed",
		})
		return
	}

	// 重新加载 advertisement 以获取关联的文件信息
	if err := databases.DB_CONN.Preload("File").First(advertisement, advertisement.ID).Error; err != nil {
		c.ctx.JSON(200, gin.H{
			"message": "create advertisement success, but failed to load file info",
			"data":    advertisement,
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create advertisement success",
		"data":    advertisement,
	})
}

func (c *BuildingAdminAdvertisementController) UpdateAdvertisement() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	var form struct {
		Title       string                  `json:"title"`
		Description string                  `json:"description"`
		Type        field.AdvertisementType `json:"type"`
		Status      field.Status            `json:"status"`
		StartTime   *time.Time              `json:"startTime"`
		EndTime     *time.Time              `json:"endTime"`
		IsPublic    *bool                   `json:"isPublic"`
		Path        string                  `json:"path"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.Type != "" {
		updates["type"] = form.Type
	}
	if form.Status != "" {
		updates["status"] = form.Status
	}
	if form.StartTime != nil {
		updates["start_time"] = form.StartTime
	}
	if form.EndTime != nil {
		updates["end_time"] = form.EndTime
	}
	updates["is_public"] = false // 强制设置为 false

	// 如果提供了新的 path
	if form.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			c.ctx.JSON(400, gin.H{
				"error":   "File not found",
				"message": err.Error(),
			})
			return
		}
		updates["file_id"] = file.ID
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取更新后的广告信息
	updatedAd, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(200, gin.H{
			"message": "update advertisement success, but failed to load updated data",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "update advertisement success",
		"data":    updatedAd,
	})
}

func (c *BuildingAdminAdvertisementController) DeleteAdvertisement() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Advertisement deleted successfully"})
}
