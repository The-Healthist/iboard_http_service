package building_admin_controllers

import (
	"net/http"
	"strings"

	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/gin-gonic/gin"
)

type BuildingAdminAuthController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminAuthController(
	ctx *gin.Context,
	container *container.ServiceContainer,
) *BuildingAdminAuthController {
	return &BuildingAdminAuthController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdminAuth returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdminAuth(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "login":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAuthController(ctx, container)
			controller.Login()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Login 建筑管理员登录
// @Summary      建筑管理员登录
// @Description  建筑管理员通过邮箱和密码登录系统
// @Tags         BuildingAdmin
// @Accept       json
// @Produce      json
// @Param        request body object true "登录信息"
// @Param        email formData string true "邮箱" example:"admin@building.com"
// @Param        password formData string true "密码" example:"password123"
// @Success      200  {object}  map[string]interface{} "包含token和登录成功消息"
// @Failure      400  {object}  map[string]interface{} "登录失败信息"
// @Failure      401  {object}  map[string]interface{} "认证失败"
// @Failure      500  {object}  map[string]interface{} "服务器错误信息"
// @Router       /building_admin/login [post]
// @Security     None
func (c *BuildingAdminAuthController) Login() {
	var loginDTO struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&loginDTO); err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	// Clean input
	loginDTO.Email = strings.TrimSpace(loginDTO.Email)
	loginDTO.Password = strings.TrimSpace(loginDTO.Password)

	// Validate credentials
	buildingAdmin, err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).GetByEmail(loginDTO.Email)
	if err != nil {
		c.Ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Validate password
	if !c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).ValidatePassword(buildingAdmin, loginDTO.Password) {
		c.Ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// Generate token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateBuildingAdminToken(buildingAdmin)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.Ctx.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"data": gin.H{
			"id":    buildingAdmin.ID,
			"email": buildingAdmin.Email,
		},
	})
}
