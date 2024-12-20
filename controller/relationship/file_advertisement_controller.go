package http_relationship_controller

import (
	"strconv"

	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileAdvertisementController interface {
	BindFile()
	UnbindFile()
	GetAdvertisementByFile()
	GetFileByAdvertisement()
}

type FileAdvertisementController struct {
	ctx     *gin.Context
	service relationship_service.InterfaceFileAdvertisementService
}

func NewFileAdvertisementController(
	ctx *gin.Context,
	service relationship_service.InterfaceFileAdvertisementService,
) InterfaceFileAdvertisementController {
	return &FileAdvertisementController{
		ctx:     ctx,
		service: service,
	}
}

func (c *FileAdvertisementController) BindFile() {
	var form struct {
		AdvertisementID uint `json:"advertisementId" binding:"required"`
		FileID          uint `json:"fileId" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	exists, err := c.service.AdvertisementExists(form.AdvertisementID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	exists, err = c.service.FileExists(form.FileID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "File not found"})
		return
	}

	if err := c.service.BindFile(form.AdvertisementID, form.FileID); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

func (c *FileAdvertisementController) UnbindFile() {
	advertisementIDStr := c.ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	if err := c.service.UnbindFile(uint(advertisementID)); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

func (c *FileAdvertisementController) GetAdvertisementByFile() {
	fileIDStr := c.ctx.Query("fileId")
	if fileIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "fileId is required"})
		return
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid fileId"})
		return
	}

	advertisement, err := c.service.GetAdvertisementByFileID(uint(fileID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch advertisement"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": advertisement})
}

func (c *FileAdvertisementController) GetFileByAdvertisement() {
	advertisementIDStr := c.ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	file, err := c.service.GetFileByAdvertisementID(uint(advertisementID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": file})
}
