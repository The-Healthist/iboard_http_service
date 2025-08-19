package http_base_controller

import (
	models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/gin-gonic/gin"
)

// AppController 应用版本控制器
type AppController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewAppController 创建控制器
func NewAppController(ctx *gin.Context, container *container.ServiceContainer) *AppController {
	return &AppController{Ctx: ctx, Container: container}
}

// HandleFuncApp 根据方法返回处理函数
func HandleFuncApp(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	// 获取应用版本配置
	case "get":
		return func(ctx *gin.Context) { NewAppController(ctx, container).Get() }
	// 更新应用版本配置
	case "update":
		return func(ctx *gin.Context) { NewAppController(ctx, container).Update() }
	default:
		return func(ctx *gin.Context) { ctx.JSON(400, gin.H{"error": "invalid method"}) }
	}
}

// Get 获取应用版本配置
// @Summary      获取应用版本配置
// @Description  获取当前应用版本配置信息，包括当前使用的版本详情
// @Tags         App
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{} "返回应用版本配置信息"
// @Failure      500  {object}  map[string]interface{} "错误信息"
// @Router       /api/app/version [get]
func (c *AppController) Get() {
	app, err := c.Container.GetService("app").(base_services.InterfaceAppService).Get()
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 确保返回的数据结构清晰
	response := gin.H{
		"data": gin.H{
			"id":               app.ID,
			"currentVersionId": app.CurrentVersionID,
			"currentVersion":   app.CurrentVersion,
			"lastCheckTime":    app.LastCheckTime,
			"updateInterval":   app.UpdateInterval,
			"autoUpdate":       app.AutoUpdate,
			"status":           app.Status,
			"createdAt":        app.CreatedAt,
			"updatedAt":        app.UpdatedAt,
		},
		"message": "Get app version config success",
	}

	c.Ctx.JSON(200, response)
}

// Update 更新应用版本配置
// @Summary      更新应用版本配置
// @Description  更新应用版本配置信息，包括设置当前使用的版本
// @Tags         App
// @Accept       json
// @Produce      json
// @Param        app body models.App true "应用版本配置信息"
// @Success      200  {object}  map[string]interface{} "返回更新后的应用版本配置信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /api/admin/app/version [put]
// @Security     BearerAuth
func (c *AppController) Update() {
	var app models.App
	if err := c.Ctx.ShouldBindJSON(&app); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "invalid form"})
		return
	}

	updatedApp, err := c.Container.GetService("app").(base_services.InterfaceAppService).Update(&app)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"message": "Update app version config success", "data": updatedApp})
}
