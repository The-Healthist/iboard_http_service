package building_admin_controllers

import (
	"net/http"
	"strings"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

type BuildingAdminAuthController struct {
	ctx                  *gin.Context
	buildingAdminService base_services.InterfaceBuildingAdminService
	jwtService           base_services.IJWTService
}

func NewBuildingAdminAuthController(
	ctx *gin.Context,
	buildingAdminService base_services.InterfaceBuildingAdminService,
	jwtService base_services.IJWTService,
) *BuildingAdminAuthController {
	return &BuildingAdminAuthController{
		ctx:                  ctx,
		buildingAdminService: buildingAdminService,
		jwtService:           jwtService,
	}
}

func (c *BuildingAdminAuthController) Login() {
	var loginDTO struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&loginDTO); err != nil {
		c.ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	// 清理输入
	loginDTO.Email = strings.TrimSpace(loginDTO.Email)
	loginDTO.Password = strings.TrimSpace(loginDTO.Password)

	// 验证凭据
	buildingAdmin, err := c.buildingAdminService.GetByEmail(loginDTO.Email)
	if err != nil {
		c.ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 验证密码
	if !c.buildingAdminService.ValidatePassword(buildingAdmin, loginDTO.Password) {
		c.ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 生成 token
	token, err := c.jwtService.GenerateBuildingAdminToken(buildingAdmin)
	if err != nil {
		c.ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.ctx.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"data": gin.H{
			"id":    buildingAdmin.ID,
			"email": buildingAdmin.Email,
		},
	})
}
