package http_relationship_controller

import (
	"strconv"

	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileNoticeController interface {
	BindFile()
	UnbindFile()
	GetNoticeByFile()
	GetFileByNotice()
}

type FileNoticeController struct {
	ctx     *gin.Context
	service relationship_service.InterfaceFileNoticeService
}

func NewFileNoticeController(
	ctx *gin.Context,
	service relationship_service.InterfaceFileNoticeService,
) InterfaceFileNoticeController {
	return &FileNoticeController{
		ctx:     ctx,
		service: service,
	}
}

func (c *FileNoticeController) BindFile() {
	var form struct {
		NoticeID uint `json:"noticeId" binding:"required"`
		FileID   uint `json:"fileId" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	// 检查通知是否存在
	exists, err := c.service.NoticeExists(form.NoticeID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Notice not found"})
		return
	}

	// 检查文件是否存在
	exists, err = c.service.FileExists(form.FileID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "File not found"})
		return
	}

	if err := c.service.BindFile(form.NoticeID, form.FileID); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

func (c *FileNoticeController) UnbindFile() {
	noticeIDStr := c.ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	if err := c.service.UnbindFile(uint(noticeID)); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

func (c *FileNoticeController) GetNoticeByFile() {
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

	notice, err := c.service.GetNoticeByFileID(uint(fileID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch notice"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": notice})
}

func (c *FileNoticeController) GetFileByNotice() {
	noticeIDStr := c.ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	file, err := c.service.GetFileByNoticeID(uint(noticeID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": file})
}
