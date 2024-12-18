package http_controller

import (
	"github.com/The-Healthist/iboard_http_service/models"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementController interface {
	Create()
	Get()
	Update()
	Delete()
}

type AdvertisementController struct {
	ctx     *gin.Context
	service services.InterfaceAdvertisementService
}

func NewAdvertisementController(
	ctx *gin.Context,
	service services.InterfaceAdvertisementService,
) InterfaceAdvertisementController {
	return &AdvertisementController{
		ctx:     ctx,
		service: service,
	}
}

func (c *AdvertisementController) Create() {
	var form struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		FileID      uint   `json:"fileId" binding:"required"`
		Duration    int    `json:"duration"`
		Display     string `json:"display" binding:"required"`
		BuildingIDs []uint `json:"buildingIds"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	advertisement := &models.Advertisement{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		FileID:      form.FileID,
		Duration:    form.Duration,
		Display:     form.Display,
		Active:      true,
	}

	if err := c.service.Create(advertisement, form.BuildingIDs); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create advertisement failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create advertisement success",
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
		FileID      uint   `json:"fileId"`
		Duration    int    `json:"duration"`
		Display     string `json:"display"`
		Active      *bool  `json:"active"`
		BuildingIDs []uint `json:"buildingIds"`
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
	if form.FileID != 0 {
		updates["file_id"] = form.FileID
	}
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
