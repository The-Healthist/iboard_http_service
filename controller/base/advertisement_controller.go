package http_base_controller

import (
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementController interface {
	Create()
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

func (c *AdvertisementController) Create() {
	var form struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		Type        field.AdvertisementType    `json:"type" binding:"required"`
		Duration    int                        `json:"duration"`
		Display     field.AdvertisementDisplay `json:"display" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   "Invalid JSON format: " + err.Error(),
			"message": "Please check your request body format",
		})
		return
	}

	advertisement := &base_models.Advertisement{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		Duration:    form.Duration,
		Display:     form.Display,
		Active:      true,
	}

	if err := c.service.Create(advertisement); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to create advertisement",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Advertisement created successfully",
		"data":    advertisement,
	})
}

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

func (c *AdvertisementController) Update() {
	var form struct {
		ID          uint   `json:"id" binding:"required"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Duration    int    `json:"duration"`
		Display     string `json:"display"`
		Active      *bool  `json:"active"`
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
	if form.Duration != 0 {
		updates["duration"] = form.Duration
	}
	if form.Display != "" {
		updates["display"] = form.Display
	}
	if form.Active != nil {
		updates["active"] = *form.Active
	}

	advertisement, err := c.service.Update(form.ID, updates)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Advertisement updated successfully",
		"data":    advertisement,
	})
}

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
