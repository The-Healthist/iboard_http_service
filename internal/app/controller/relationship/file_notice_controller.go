package http_relationship_controller

import (
	"strconv"

	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileNoticeController interface {
	BindFile()
	UnbindFile()
	GetNoticeByFile()
	GetFileByNotice()
}

type FileNoticeController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileNoticeController(ctx *gin.Context, container *container.ServiceContainer) *FileNoticeController {
	return &FileNoticeController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFileNotice returns a gin.HandlerFunc for the specified method
func HandleFuncFileNotice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewFileNoticeController(ctx, container)
		switch method {
		case "bindFile":
			controller.BindFile()
		case "unbindFile":
			controller.UnbindFile()
		case "getNoticeByFile":
			controller.GetNoticeByFile()
		case "getFileByNotice":
			controller.GetFileByNotice()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *FileNoticeController) getService() relationship_service.InterfaceFileNoticeService {
	return c.Container.GetService("fileNotice").(relationship_service.InterfaceFileNoticeService)
}

// 1. BindFile 绑定文件到通知
// @Summary      绑定文件到通知
// @Description  将一个文件绑定到一个通知
// @Tags         File-Notice
// @Accept       json
// @Produce      json
// @Param        bindInfo body object true "绑定信息"
// @Param        noticeId body uint true "通知ID" example:"1"
// @Param        fileId body uint true "文件ID" example:"2"
// @Success      200  {object}  map[string]interface{} "绑定成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误"
// @Failure      404  {object}  map[string]interface{} "通知或文件不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-notice/bind [post]
// @Security     BearerAuth
func (c *FileNoticeController) BindFile() {
	var form struct {
		NoticeID uint `json:"noticeId" binding:"required"`
		FileID   uint `json:"fileId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	// 检查通知是否存在
	exists, err := c.getService().NoticeExists(form.NoticeID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Notice not found"})
		return
	}

	// 检查文件是否存在
	exists, err = c.getService().FileExists(form.FileID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "File not found"})
		return
	}

	if err := c.getService().BindFile(form.NoticeID, form.FileID); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

// 2. UnbindFile 解绑文件与通知
// @Summary      解绑文件与通知
// @Description  解除一个通知与其绑定的文件的关系
// @Tags         File-Notice
// @Accept       json
// @Produce      json
// @Param        noticeId query int true "通知ID" example:"1"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "无效的通知ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-notice/unbind [get]
// @Security     BearerAuth
func (c *FileNoticeController) UnbindFile() {
	noticeIDStr := c.Ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	if err := c.getService().UnbindFile(uint(noticeID)); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

// 3. GetNoticeByFile 根据文件获取通知
// @Summary      根据文件获取通知
// @Description  获取与指定文件关联的通知信息
// @Tags         File-Notice
// @Accept       json
// @Produce      json
// @Param        fileId query int true "文件ID" example:"1"
// @Success      200  {object}  map[string]interface{} "通知信息"
// @Failure      400  {object}  map[string]interface{} "无效的文件ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-notice/notice [get]
// @Security     BearerAuth
func (c *FileNoticeController) GetNoticeByFile() {
	fileIDStr := c.Ctx.Query("fileId")
	if fileIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "fileId is required"})
		return
	}

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid fileId"})
		return
	}

	notice, err := c.getService().GetNoticeByFileID(uint(fileID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch notice"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": notice})
}

// 4. GetFileByNotice 根据通知获取文件
// @Summary      根据通知获取文件
// @Description  获取与指定通知关联的文件信息
// @Tags         File-Notice
// @Accept       json
// @Produce      json
// @Param        noticeId query int true "通知ID" example:"1"
// @Success      200  {object}  map[string]interface{} "文件信息"
// @Failure      400  {object}  map[string]interface{} "无效的通知ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-notice/file [get]
// @Security     BearerAuth
func (c *FileNoticeController) GetFileByNotice() {
	noticeIDStr := c.Ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	file, err := c.getService().GetFileByNoticeID(uint(noticeID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": file})
}
