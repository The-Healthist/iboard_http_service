package http_controller

import (
	"github.com/The-Healthist/iboard_http_service/models"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceFileController interface {
	Create()
	Get()
	Update()
	Delete()
}

type FileController struct {
	ctx     *gin.Context
	service services.InterfaceFileService
}

func NewFileController(
	ctx *gin.Context,
	service services.InterfaceFileService,
) InterfaceFileController {
	return &FileController{
		ctx:     ctx,
		service: service,
	}
}

func (c *FileController) Create() {
	var form struct {
		Path         string `json:"path" binding:"required"`
		Size         int64  `json:"size" binding:"required"`
		MimeType     string `json:"mimeType" binding:"required"`
		Oss          string `json:"oss" binding:"required"`
		Uploader     string `json:"uploader" binding:"required"`
		UploaderType string `json:"uploaderType" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	file := &models.File{
		Path:         form.Path,
		Size:         form.Size,
		MimeType:     form.MimeType,
		Oss:          form.Oss,
		Uploader:     form.Uploader,
		UploaderType: form.UploaderType,
	}

	if err := c.service.Create(file); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create file failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create file success",
		"data":    file,
	})
}

func (c *FileController) Get() {
	var searchQuery struct {
		MimeType     string `form:"mimeType"`
		Oss          string `form:"oss"`
		UploaderType string `form:"uploaderType"`
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

	files, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": paginationResult,
	})
}

func (c *FileController) Update() {
	var form struct {
		ID           uint   `json:"id" binding:"required"`
		Path         string `json:"path"`
		Size         int64  `json:"size"`
		MimeType     string `json:"mimeType"`
		Oss          string `json:"oss"`
		Uploader     string `json:"uploader"`
		UploaderType string `json:"uploaderType"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Path != "" {
		updates["path"] = form.Path
	}
	if form.Size != 0 {
		updates["size"] = form.Size
	}
	if form.MimeType != "" {
		updates["mime_type"] = form.MimeType
	}
	if form.Oss != "" {
		updates["oss"] = form.Oss
	}
	if form.Uploader != "" {
		updates["uploader"] = form.Uploader
	}
	if form.UploaderType != "" {
		updates["uploader_type"] = form.UploaderType
	}

	if err := c.service.Update(form.ID, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "update file success"})
}

func (c *FileController) Delete() {
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

	c.ctx.JSON(200, gin.H{"message": "delete file success"})
}
