package http_base_controller

import (
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceNoticeController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
}

type NoticeController struct {
	ctx     *gin.Context
	service base_services.InterfaceNoticeService
}

func NewNoticeController(
	ctx *gin.Context,
	service base_services.InterfaceNoticeService,
) InterfaceNoticeController {
	return &NoticeController{
		ctx:     ctx,
		service: service,
	}
}

func (c *NoticeController) Create() {
	var form struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type"`
		FileID      *uint  `json:"fileId"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	notice := &base_models.Notice{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		FileID:      form.FileID,
	}

	if err := c.service.Create(notice); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create notice failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create notice success",
		"data":    notice,
	})
}

func (c *NoticeController) Get() {
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
		Desc:     true,
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

	notices, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": paginationResult,
	})
}

func (c *NoticeController) Update() {
	var form struct {
		ID          uint   `json:"id" binding:"required"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		FileID      *uint  `json:"fileId"`
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
	if form.FileID != nil {
		updates["file_id"] = *form.FileID
	}

	notice, err := c.service.Update(form.ID, updates)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "update notice success",
		"data":    notice,
	})
}

func (c *NoticeController) Delete() {
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

	c.ctx.JSON(200, gin.H{"message": "delete notice success"})
}

func (c *NoticeController) GetOne() {
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	notice, err := c.service.GetByID(uint(id))
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": notice})
}
