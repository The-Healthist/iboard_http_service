package http_base_controller

import (
	"strconv"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceBuildingController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
	Login()
	GetBuildingAdvertisements()
	GetBuildingNotices()
}

type BuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingController(ctx *gin.Context, container *container.ServiceContainer) *BuildingController {
	return &BuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetOne()
		}
	case "login":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Login()
		}
	case "getBuildingAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetBuildingAdvertisements()
		}
	case "getBuildingNotices":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetBuildingNotices()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Create creates a new building
func (c *BuildingController) Create() {
	var form struct {
		Name     string `json:"name" binding:"required"`
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
		Remark   string `json:"remark"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	building := &base_models.Building{
		Name:     form.Name,
		IsmartID: form.IsmartID,
		Password: form.Password,
		Remark:   form.Remark,
	}

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Create(building); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create building success",
		"data":    building,
	})
}

func (c *BuildingController) Get() {
	var searchQuery struct {
		Search string `form:"search"`
	}
	if err := c.Ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
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

	buildings, paginationResult, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       buildings,
		"pagination": paginationResult,
	})
}

func (c *BuildingController) Update() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Name     string `json:"name"`
		IsmartID string `json:"ismartId"`
		Password string `json:"password"`
		Remark   string `json:"remark"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Name != "" {
		updates["name"] = form.Name
	}
	if form.IsmartID != "" {
		updates["ismart_id"] = form.IsmartID
	}
	if form.Password != "" {
		updates["password"] = form.Password
	}
	if form.Remark != "" {
		updates["remark"] = form.Remark
	}

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update building success"})
}

func (c *BuildingController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get buildings to delete
	var buildings []base_models.Building
	if err := tx.Where("id IN ?", form.IDs).Find(&buildings).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get buildings",
			"message": err.Error(),
		})
		return
	}

	// 2. Remove advertisement associations
	if err := tx.Exec("DELETE FROM advertisement_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind advertisements",
			"message": err.Error(),
		})
		return
	}

	// 3. Remove notice associations
	if err := tx.Exec("DELETE FROM notice_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind notices",
			"message": err.Error(),
		})
		return
	}

	// 4. Remove admin associations
	if err := tx.Exec("DELETE FROM building_admins_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind admins",
			"message": err.Error(),
		})
		return
	}

	// 5. Delete buildings
	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Delete(form.IDs); err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete buildings",
			"message": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete building success"})
}

func (c *BuildingController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building ID",
			"message": "Please check the ID format",
		})
		return
	}

	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get building",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get building success",
		"data":    building,
	})
}

func (c *BuildingController) Login() {
	var form struct {
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByCredentials(form.IsmartID, form.Password)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateBuildingToken(building)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to generate token",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Login success",
		"data": gin.H{
			"id":       building.ID,
			"name":     building.Name,
			"ismartId": building.IsmartID,
			"remark":   building.Remark,
		},
		"token": token,
	})
}

func (c *BuildingController) GetBuildingAdvertisements() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	advertisements, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetBuildingAdvertisements(buildingId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get advertisements",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get advertisements success",
		"data":    advertisements,
	})
}

func (c *BuildingController) GetBuildingNotices() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	notices, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetBuildingNotices(buildingId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get notices",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get notices success",
		"data":    notices,
	})
}
