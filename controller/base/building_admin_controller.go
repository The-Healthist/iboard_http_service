package http_base_controller

import (
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type InterfaceBuildingAdminController interface {
	Create()
	Get()
	Update()
	Delete()
}

type BuildingAdminController struct {
	ctx     *gin.Context
	service base_services.InterfaceBuildingAdminService
}

func NewBuildingAdminController(
	ctx *gin.Context,
	service base_services.InterfaceBuildingAdminService,
) InterfaceBuildingAdminController {
	return &BuildingAdminController{
		ctx:     ctx,
		service: service,
	}
}

func (c *BuildingAdminController) Create() {
	var form struct {
		Email    string `json:"email"      binding:"required"`
		Password string `json:"password"   binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		c.ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "password encryption failed",
		})
		return
	}

	buildingAdmin := &base_models.BuildingAdmin{
		Email:    form.Email,
		Password: string(hashedPassword),
		Status:   true,
	}

	if err := c.service.Create(buildingAdmin); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building admin failed",
		})
		return
	}

	buildingAdmin.Password = "" // Don't return password
	c.ctx.JSON(200, gin.H{
		"message": "create building admin success",
		"data":    buildingAdmin,
	})
}

func (c *BuildingAdminController) Get() {
	var searchQuery struct {
		BuildingID string `form:"buildingId"`
		Status     *bool  `form:"status"`
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

	buildingAdmins, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       buildingAdmins,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminController) Update() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Password string `json:"password"`
		Status   *bool  `json:"status                                      "`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if form.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			c.ctx.JSON(500, gin.H{"error": "password encryption failed"})
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if form.Status != nil {
		updates["status"] = *form.Status
	}

	if err := c.service.Update(form.ID, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "update building admin success"})
}

func (c *BuildingAdminController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Delete(form.IDs); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete building admin success"})
}
