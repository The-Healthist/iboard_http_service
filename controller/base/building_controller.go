package http_base_controller

import (
	"strconv"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
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
	ctx        *gin.Context
	service    base_services.InterfaceBuildingService
	jwtService *base_services.IJWTService
}

func NewBuildingController(
	ctx *gin.Context,
	service base_services.InterfaceBuildingService,
	jwtService *base_services.IJWTService,
) InterfaceBuildingController {
	return &BuildingController{
		ctx:        ctx,
		service:    service,
		jwtService: jwtService,
	}
}

// 1,create
func (c *BuildingController) Create() {
	var form struct {
		Name     string `json:"name" binding:"required"`
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
		Remark   string `json:"remark"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
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

	if err := c.service.Create(building); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create building success",
		"data":    building,
	})
}

// 2,get
func (c *BuildingController) Get() {
	var searchQuery struct {
		Search string `form:"search"`
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

	buildings, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       buildings,
		"pagination": paginationResult,
	})
}

// 3,update
func (c *BuildingController) Update() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Name     string `json:"name"`
		IsmartID string `json:"ismartId"`
		Password string `json:"password"`
		Remark   string `json:"remark"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
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

	if err := c.service.Update(form.ID, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "update building success"})
}

// 4,delete
func (c *BuildingController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 开启事务
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取要删除的建筑物信息
	var buildings []base_models.Building
	if err := tx.Where("id IN ?", form.IDs).Find(&buildings).Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to get buildings",
			"message": err.Error(),
		})
		return
	}

	// 2. 解除与广告的关联
	if err := tx.Exec("DELETE FROM advertisement_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to unbind advertisements",
			"message": err.Error(),
		})
		return
	}

	// 3. 解除与通知的关联
	if err := tx.Exec("DELETE FROM notice_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to unbind notices",
			"message": err.Error(),
		})
		return
	}

	// 4. 解除与管理员的关联
	if err := tx.Exec("DELETE FROM building_admins_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to unbind admins",
			"message": err.Error(),
		})
		return
	}

	// 5. 删除建筑物
	if err := tx.Delete(&base_models.Building{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to delete buildings",
			"message": err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.ctx.JSON(400, gin.H{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete building success"})
}

// 5,get one
func (c *BuildingController) GetOne() {
	if c.jwtService == nil {
		c.ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   "Invalid building ID",
			"message": "Please check the ID format",
		})
		return
	}

	building, err := c.service.GetByID(uint(id))
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get building",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Get building success",
		"data":    building,
	})
}

// 6,login
func (c *BuildingController) Login() {
	var form struct {
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	building, err := c.service.GetByCredentials(form.IsmartID, form.Password)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	if c.jwtService == nil {
		c.ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	token, err := (*c.jwtService).GenerateBuildingToken(building)
	if err != nil {
		c.ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to generate token",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
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

// 7,get building advertisements
func (c *BuildingController) GetBuildingAdvertisements() {
	claims, exists := c.ctx.Get("claims")
	if !exists {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	advertisements, err := c.service.GetBuildingAdvertisements(buildingId)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get advertisements",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Get advertisements success",
		"data":    advertisements,
	})
}

// 8,get building notices
func (c *BuildingController) GetBuildingNotices() {
	claims, exists := c.ctx.Get("claims")
	if !exists {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	notices, err := c.service.GetBuildingNotices(buildingId)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get notices",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Get notices success",
		"data":    notices,
	})
}
