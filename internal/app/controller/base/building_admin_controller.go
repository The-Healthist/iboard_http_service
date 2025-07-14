package http_base_controller

import (
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type InterfaceBuildingAdminController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
}

type BuildingAdminController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminController(ctx *gin.Context, container *container.ServiceContainer) *BuildingAdminController {
	return &BuildingAdminController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdmin returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdmin(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// 1.Create 创建建筑管理员
// @Summary      创建建筑管理员
// @Description  创建一个新的建筑管理员账户
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        buildingAdmin body object true "建筑管理员信息"
// @Param        email formData string true "邮箱" example:"admin@building.com"
// @Param        password formData string true "密码" example:"password123"
// @Param        status formData string true "状态" example:"active"
// @Success      200  {object}  map[string]interface{} "返回创建的建筑管理员信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      500  {object}  map[string]interface{} "服务器错误"
// @Router       /admin/building_admin [post]
// @Security     BearerAuth
func (c *BuildingAdminController) Create() {
	var form struct {
		Email    string       `json:"email"    binding:"required" example:"admin@building.com"`
		Password string       `json:"password" binding:"required" example:"password123"`
		Status   field.Status `json:"status"   binding:"required" example:"active"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Validate status
	if !field.IsValidStatus(string(form.Status)) {
		c.Ctx.JSON(400, gin.H{
			"error":   "invalid status",
			"message": "status must be one of: active, inactive, pending",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "password encryption failed",
		})
		return
	}

	buildingAdmin := &base_models.BuildingAdmin{
		Email:    form.Email,
		Password: string(hashedPassword),
		Status:   form.Status,
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Create(buildingAdmin); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building admin failed",
		})
		return
	}

	buildingAdmin.Password = "" // Don't return password
	c.Ctx.JSON(200, gin.H{
		"message": "create building admin success",
		"data":    buildingAdmin,
	})
}

// 2.Get 获取建筑管理员列表
// @Summary      获取建筑管理员列表
// @Description  根据查询条件获取建筑管理员列表
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        buildingId query string false "建筑ID" example:"1"
// @Param        status query string false "状态" example:"active"
// @Param        pageSize query int false "每页数量" default(10)
// @Param        pageNum query int false "页码" default(1)
// @Param        desc query bool false "是否降序" default(false)
// @Success      200  {object}  map[string]interface{} "返回建筑管理员列表和分页信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/building_admin [get]
// @Security     BearerAuth
func (c *BuildingAdminController) Get() {
	var searchQuery struct {
		BuildingID string        `form:"buildingId" example:"1"`
		Status     *field.Status `form:"status" example:"active"`
	}
	if err := c.Ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pagination := struct {
		PageSize int  `form:"pageSize" example:"10"`
		PageNum  int  `form:"pageNum" example:"1"`
		Desc     bool `form:"desc" example:"false"`
	}{
		PageSize: 10,
		PageNum:  1,
		Desc:     false,
	}

	if err := c.Ctx.ShouldBindQuery(&pagination); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	queryMap := utils.StructToMap(searchQuery)
	paginationMap := map[string]interface{}{
		"pageSize": pagination.PageSize,
		"pageNum":  pagination.PageNum,
		"desc":     pagination.Desc,
	}

	buildingAdmins, paginationResult, err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       buildingAdmins,
		"pagination": paginationResult,
	})
}

// 3.Update 更新建筑管理员
// @Summary      更新建筑管理员
// @Description  更新建筑管理员信息
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        buildingAdmin body object true "建筑管理员更新信息"
// @Param        id formData uint true "建筑管理员ID" example:"1"
// @Param        password formData string false "密码" example:"newpassword123"
// @Param        status formData string false "状态" example:"inactive"
// @Success      200  {object}  map[string]interface{} "更新成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      500  {object}  map[string]interface{} "服务器错误"
// @Router       /admin/building_admin [put]
// @Security     BearerAuth
func (c *BuildingAdminController) Update() {
	var form struct {
		ID       uint          `json:"id" binding:"required" example:"1"`
		Password string        `json:"password" example:"newpassword123"`
		Status   *field.Status `json:"status" example:"inactive"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if form.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "password encryption failed"})
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if form.Status != nil {
		updates["status"] = *form.Status
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update building admin success"})
}

// 4.Delete 删除建筑管理员
// @Summary      删除建筑管理员
// @Description  删除一个或多个建筑管理员
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        ids body object true "建筑管理员ID列表"
// @Param        ids.ids body []uint true "建筑管理员ID数组" example:"[1,2,3]"
// @Success      200  {object}  map[string]interface{} "删除成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/building_admin [delete]
// @Security     BearerAuth
func (c *BuildingAdminController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required" example:"[1,2,3]"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete building admin success"})
}

// 5.GetOne 获取单个建筑管理员
// @Summary      获取单个建筑管理员
// @Description  根据ID获取建筑管理员详细信息
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        id path int true "建筑管理员ID" example:"1"
// @Success      200  {object}  map[string]interface{} "返回建筑管理员详细信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      500  {object}  map[string]interface{} "服务器错误"
// @Router       /admin/building_admin/{id} [get]
// @Security     BearerAuth
func (c *BuildingAdminController) GetOne() {
	if c.Container.GetService("jwt") == nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building admin ID",
			"message": "Please check the ID format",
		})
		return
	}

	buildingAdmin, err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get building admin",
		})
		return
	}

	buildingAdmin.Password = "" // Don't return password
	c.Ctx.JSON(200, gin.H{
		"message": "Get building admin success",
		"data":    buildingAdmin,
	})
}
