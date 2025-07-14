package http_relationship_controller

import (
	"strconv"

	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceFileAdvertisementController interface {
	BindFile()
	UnbindFile()
	GetAdvertisementByFile()
	GetFileByAdvertisement()
}

type FileAdvertisementController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileAdvertisementController(ctx *gin.Context, container *container.ServiceContainer) *FileAdvertisementController {
	return &FileAdvertisementController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFileAdvertisement returns a gin.HandlerFunc for the specified method
func HandleFuncFileAdvertisement(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewFileAdvertisementController(ctx, container)
		switch method {
		case "bindFile":
			controller.BindFile()
		case "unbindFile":
			controller.UnbindFile()
		case "getAdvertisementByFile":
			controller.GetAdvertisementByFile()
		case "getFileByAdvertisement":
			controller.GetFileByAdvertisement()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *FileAdvertisementController) getService() relationship_service.InterfaceFileAdvertisementService {
	return c.Container.GetService("fileAdvertisement").(relationship_service.InterfaceFileAdvertisementService)
}

// 1. BindFile 绑定文件到广告
// @Summary      绑定文件到广告
// @Description  将一个文件绑定到一个广告
// @Tags         File-Advertisement
// @Accept       json
// @Produce      json
// @Param        bindInfo body object true "绑定信息"
// @Param        advertisementId body uint true "广告ID" example:"1"
// @Param        fileId body uint true "文件ID" example:"2"
// @Success      200  {object}  map[string]interface{} "绑定成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误"
// @Failure      404  {object}  map[string]interface{} "广告或文件不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-advertisement/bind [post]
// @Security     BearerAuth
func (c *FileAdvertisementController) BindFile() {
	var form struct {
		AdvertisementID uint `json:"advertisementId" binding:"required"`
		FileID          uint `json:"fileId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters"})
		return
	}

	// 检查广告是否存在
	exists, err := c.getService().AdvertisementExists(form.AdvertisementID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Advertisement not found"})
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

	if err := c.getService().BindFile(form.AdvertisementID, form.FileID); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File bound successfully"})
}

// 2. UnbindFile 解绑文件与广告
// @Summary      解绑文件与广告
// @Description  解除一个广告与其绑定的文件的关系
// @Tags         File-Advertisement
// @Accept       json
// @Produce      json
// @Param        advertisementId query int true "广告ID" example:"1"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "无效的广告ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-advertisement/unbind [get]
// @Security     BearerAuth
func (c *FileAdvertisementController) UnbindFile() {
	advertisementIDStr := c.Ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	if err := c.getService().UnbindFile(uint(advertisementID)); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "File unbound successfully"})
}

// 3. GetAdvertisementByFile 根据文件获取广告
// @Summary      根据文件获取广告
// @Description  获取与指定文件关联的广告信息
// @Tags         File-Advertisement
// @Accept       json
// @Produce      json
// @Param        fileId query int true "文件ID" example:"1"
// @Success      200  {object}  map[string]interface{} "广告信息"
// @Failure      400  {object}  map[string]interface{} "无效的文件ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-advertisement/advertisement [get]
// @Security     BearerAuth
func (c *FileAdvertisementController) GetAdvertisementByFile() {
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

	advertisement, err := c.getService().GetAdvertisementByFileID(uint(fileID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch advertisement"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": advertisement})
}

// 4. GetFileByAdvertisement 根据广告获取文件
// @Summary      根据广告获取文件
// @Description  获取与指定广告关联的文件信息
// @Tags         File-Advertisement
// @Accept       json
// @Produce      json
// @Param        advertisementId query int true "广告ID" example:"1"
// @Success      200  {object}  map[string]interface{} "文件信息"
// @Failure      400  {object}  map[string]interface{} "无效的广告ID"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/file-advertisement/file [get]
// @Security     BearerAuth
func (c *FileAdvertisementController) GetFileByAdvertisement() {
	advertisementIDStr := c.Ctx.Query("advertisementId")
	if advertisementIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	advertisementID, err := strconv.ParseUint(advertisementIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	file, err := c.getService().GetFileByAdvertisementID(uint(advertisementID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch file"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": file})
}
