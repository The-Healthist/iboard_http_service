package http_base_controller

import (
    "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
    "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
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
    // 1.Get 获取 App 版本
    case "get":
        return func(ctx *gin.Context) { NewAppController(ctx, container).Get() }
    // 2.Update 更新 App 版本
    case "update":
        return func(ctx *gin.Context) { NewAppController(ctx, container).Update() }
    default:
        return func(ctx *gin.Context) { ctx.JSON(400, gin.H{"error": "invalid method"}) }
    }
}

// 1.Get 获取App版本
func (c *AppController) Get() {
    app, err := c.Container.GetService("app").(base_services.InterfaceAppService).Get()
    if err != nil {
        c.Ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.Ctx.JSON(200, gin.H{"data": app})
}

// 2.Update 更新App版本
func (c *AppController) Update() {
    var form struct {
        Version string `json:"version" binding:"required"`
    }
    if err := c.Ctx.ShouldBindJSON(&form); err != nil {
        c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "invalid form"})
        return
    }
    app, err := c.Container.GetService("app").(base_services.InterfaceAppService).Update(form.Version)
    if err != nil {
        c.Ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.Ctx.JSON(200, gin.H{"message": "update success", "data": app})
}


