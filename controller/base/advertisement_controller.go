package http_base_controller

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

type AdvertisementController struct {
	ctx        *gin.Context
	service    base_services.InterfaceAdvertisementService
	jwtService *base_services.IJWTService
}

func NewAdvertisementController(
	ctx *gin.Context,
	service base_services.InterfaceAdvertisementService,
	jwtService *base_services.IJWTService,
) InterfaceAdvertisementController {
	return &AdvertisementController{
		ctx:        ctx,
		service:    service,
		jwtService: jwtService,
	}
}

// 1,create
func (c *AdvertisementController) Create() {
	var form struct {
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

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Set default values for startTime, endTime, and status
	now := time.Now()
	startTime := now
	endTime := now.AddDate(1, 0, 0) // 1 year from now
	status := field.Status("active")

	if form.StartTime != nil {
		startTime = *form.StartTime
	}
	if form.EndTime != nil {
		endTime = *form.EndTime
	}
	if form.Status != "" {
		status = form.Status
	}

	// 如果提供了 path，查找对应的文件
	var fileID *uint
	if form.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			c.ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID
	}

	advertisement := &base_models.Advertisement{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		Status:      status,
		Duration:    form.Duration,
		StartTime:   startTime,
		EndTime:     endTime,
		Display:     form.Display,
		FileID:      fileID,
		IsPublic:    form.IsPublic,
	}

	if err := c.service.Create(advertisement); err != nil {
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

// 2,createMany
func (c *AdvertisementController) CreateMany() {
	var forms []struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		Type        field.AdvertisementType    `json:"type"`
		Status      field.Status               `json:"status"`
		Duration    int                        `json:"duration"`
		StartTime   *time.Time                 `json:"startTime"`
		EndTime     *time.Time                 `json:"endTime"`
		Display     field.AdvertisementDisplay `json:"display"`
		FileID      *uint                      `json:"fileId"`
		IsPublic    bool                       `json:"isPublic"`
	}

	if err := c.ctx.ShouldBindJSON(&forms); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var advertisements []*base_models.Advertisement
	for _, form := range forms {
		advertisement := &base_models.Advertisement{
			Title:       form.Title,
			Description: form.Description,
			Type:        form.Type,
			Status:      form.Status,
			Duration:    form.Duration,
			StartTime:   *form.StartTime,
			EndTime:     *form.EndTime,
			Display:     form.Display,
			FileID:      form.FileID,
			IsPublic:    form.IsPublic,
		}
		advertisements = append(advertisements, advertisement)
	}

	for _, advertisement := range advertisements {
		if err := c.service.Create(advertisement); err != nil {
			c.ctx.JSON(400, gin.H{
				"error":         err.Error(),
				"message":       "create advertisement failed",
				"advertisement": advertisement,
			})
			return
		}
	}

	c.ctx.JSON(200, gin.H{
		"message": "create advertisements success",
		"data":    advertisements,
	})
}

// 3,get
func (c *AdvertisementController) Get() {
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

	advertisements, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       advertisements,
		"pagination": paginationResult,
	})
}

// 4,update
func (c *AdvertisementController) Update() {
	var form struct {
		ID          uint                       `json:"id" binding:"required"`
		Title       *string                    `json:"title"`
		Description *string                    `json:"description"`
		Type        field.AdvertisementType    `json:"type"`
		Status      field.Status               `json:"status"`
		Duration    *int                       `json:"duration"`
		StartTime   *time.Time                 `json:"startTime"`
		EndTime     *time.Time                 `json:"endTime"`
		Display     field.AdvertisementDisplay `json:"display"`
		FileID      *uint                      `json:"fileId"`
		IsPublic    *bool                      `json:"isPublic"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get existing advertisement to check current values
	existingAd, err := c.service.GetByID(form.ID)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Advertisement not found"})
		return
	}

	updates := map[string]interface{}{}
	if form.Title != nil {
		updates["title"] = *form.Title
	}
	if form.Description != nil {
		updates["description"] = *form.Description
	}
	if form.Type != "" {
		updates["type"] = form.Type
	}
	if form.Status != "" {
		updates["status"] = form.Status
	} else if existingAd.Status == "" {
		updates["status"] = field.Status("active")
	}
	if form.Duration != nil {
		updates["duration"] = *form.Duration
	}
	if form.StartTime != nil {
		updates["start_time"] = *form.StartTime
	} else if existingAd.StartTime.IsZero() {
		updates["start_time"] = time.Now()
	}
	if form.EndTime != nil {
		updates["end_time"] = *form.EndTime
	} else if existingAd.EndTime.IsZero() {
		updates["end_time"] = time.Now().AddDate(1, 0, 0)
	}
	if form.Display != "" {
		updates["display"] = form.Display
	}
	if form.FileID != nil {
		updates["file_id"] = *form.FileID
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}

	advertisement, err := c.service.Update(form.ID, updates)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "update advertisement success",
		"data":    advertisement,
	})
}

// 5,delete
func (c *AdvertisementController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Delete(form.IDs); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete advertisement success"})
}

// 6,getOne
func (c *AdvertisementController) GetOne() {
	// First verify JWT token
	if c.jwtService == nil {
		c.ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   "Invalid advertisement ID",
			"message": "Please check the ID format",
		})
		return
	}

	advertisement, err := c.service.GetByID(uint(id))
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get advertisement",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Get advertisement success",
		"data":    advertisement,
	})
}
