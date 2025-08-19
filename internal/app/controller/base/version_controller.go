package http_base_controller

import (
	"strconv"
	"strings"

	models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/gin-gonic/gin"
)

// VersionController 版本控制器
type VersionController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewVersionController 创建控制器
func NewVersionController(ctx *gin.Context, container *container.ServiceContainer) *VersionController {
	return &VersionController{Ctx: ctx, Container: container}
}

// HandleFuncVersion 根据方法返回处理函数
func HandleFuncVersion(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).Create() }
	case "getList":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).GetList() }
	case "getOne":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).GetOne() }
	case "update":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).Update() }
	case "delete":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).Delete() }
	case "getActive":
		return func(ctx *gin.Context) { NewVersionController(ctx, container).GetActive() }
	default:
		return func(ctx *gin.Context) { ctx.JSON(400, gin.H{"error": "invalid method"}) }
	}
}

// Create 创建版本
// @Summary      创建新版本
// @Description  创建新的应用版本
// @Tags         Version
// @Accept       json
// @Produce      json
// @Param        version body models.Version true "版本信息"
// @Success      200  {object}  map[string]interface{} "返回创建的版本信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/version [post]
// @Security     BearerAuth
func (c *VersionController) Create() {
	var version models.Version
	if err := c.Ctx.ShouldBindJSON(&version); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "invalid form"})
		return
	}

	// Validate APK download URL if provided
	if version.DownloadUrl != "" {
		if !strings.Contains(version.DownloadUrl, ".apk") {
			c.Ctx.JSON(400, gin.H{"error": "Download URL must be an APK file", "message": "invalid download url"})
			return
		}
	}

	createdVersion, err := c.Container.GetService("version").(base_services.InterfaceVersionService).Create(&version)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "Create version success", "data": createdVersion})
}

// GetList 获取版本列表
// @Summary      获取版本列表
// @Description  分页获取版本列表
// @Tags         Version
// @Accept       json
// @Produce      json
// @Param        page query int false "页码" default(1)
// @Param        pageSize query int false "每页数量" default(10)
// @Success      200  {object}  map[string]interface{} "返回版本列表"
// @Failure      500  {object}  map[string]interface{} "错误信息"
// @Router       /admin/versions [get]
// @Security     BearerAuth
func (c *VersionController) GetList() {
	page, _ := strconv.Atoi(c.Ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.Ctx.DefaultQuery("pageSize", "10"))

	versions, total, err := c.Container.GetService("version").(base_services.InterfaceVersionService).GetList(page, pageSize)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data": gin.H{
			"versions": versions,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
		"message": "Get version list success",
	})
}

// GetOne 获取单个版本
// @Summary      获取单个版本
// @Description  根据ID获取版本信息
// @Tags         Version
// @Accept       json
// @Produce      json
// @Param        id path int true "版本ID"
// @Success      200  {object}  map[string]interface{} "返回版本信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/version/{id} [get]
// @Security     BearerAuth
func (c *VersionController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	version, err := c.Container.GetService("version").(base_services.InterfaceVersionService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": version, "message": "Get version success"})
}

// Update 更新版本
// @Summary      更新版本
// @Description  更新版本信息
// @Tags         Version
// @Accept       json
// @Produce      json
// @Param        version body models.Version true "版本信息"
// @Success      200  {object}  map[string]interface{} "返回更新后的版本信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/version [put]
// @Security     BearerAuth
func (c *VersionController) Update() {
	var version models.Version
	if err := c.Ctx.ShouldBindJSON(&version); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "invalid form"})
		return
	}

	updatedVersion, err := c.Container.GetService("version").(base_services.InterfaceVersionService).Update(&version)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "Update version success", "data": updatedVersion})
}

// Delete 删除版本
// @Summary      删除版本
// @Description  根据ID删除版本
// @Tags         Version
// @Accept       json
// @Produce      json
// @Param        id path int true "版本ID"
// @Success      200  {object}  map[string]interface{} "删除成功"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/version/{id} [delete]
// @Security     BearerAuth
func (c *VersionController) Delete() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	err = c.Container.GetService("version").(base_services.InterfaceVersionService).Delete(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "Delete version success"})
}

// GetActive 获取活跃版本列表
// @Summary      获取活跃版本列表
// @Description  获取所有状态为活跃的版本
// @Tags         Version
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{} "返回活跃版本列表"
// @Failure      500  {object}  map[string]interface{} "错误信息"
// @Router       /admin/versions/active [get]
// @Security     BearerAuth
func (c *VersionController) GetActive() {
	versions, err := c.Container.GetService("version").(base_services.InterfaceVersionService).GetActiveVersions()
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": versions, "message": "Get active versions success"})
}
