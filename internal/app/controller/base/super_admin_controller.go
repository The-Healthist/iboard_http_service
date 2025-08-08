package http_base_controller

import (
	"strconv"
	"strings"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// SuperAdminLoginRequest 超级管理员登录请求
type SuperAdminLoginRequest struct {
	Email    string `json:"email" binding:"required" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// SuperAdminCreateRequest 创建超级管理员请求
type SuperAdminCreateRequest struct {
	Email    string `json:"email" binding:"required" example:"newadmin@example.com"`
	Password string `json:"password" binding:"required" example:"newpassword123"`
}

// SuperAdminResetPasswordRequest 重置密码请求
type SuperAdminResetPasswordRequest struct {
	ID       uint   `json:"id" binding:"required" example:"2"`
	Password string `json:"password" binding:"required" example:"newpassword123"`
}

// SuperAdminChangePasswordRequest 修改密码请求
type SuperAdminChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required" example:"newpassword123"`
}

// SuperAdminDeleteRequest 删除超级管理员请求
type SuperAdminDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required" example:"[2,3]"`
}

type InterfaceSuperAdminController interface {
	Login()
	GetSuperAdmins()
	CreateSuperAdmin()
	DeleteSuperAdmin()
	ResetPassword()
	ChangePassword()
	GetOne()
}

type SuperAdminController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewSuperAdminController(ctx *gin.Context, container *container.ServiceContainer) *SuperAdminController {
	return &SuperAdminController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncSuperAdmin returns a gin.HandlerFunc for the specified method
func HandleFuncSuperAdmin(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "login":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.Login()
		}
	case "getSuperAdmins":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.GetSuperAdmins()
		}
	case "createSuperAdmin":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.CreateSuperAdmin()
		}
	case "deleteSuperAdmin":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.DeleteSuperAdmin()
		}
	case "resetPassword":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.ResetPassword()
		}
	case "changePassword":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.ChangePassword()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// 1.Login 超级管理员登录
// @Summary      超级管理员登录
// @Description  超级管理员通过邮箱和密码登录系统
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        request body SuperAdminLoginRequest true "登录信息"
// @Success      200  {object}  map[string]interface{} "包含token和登录成功消息"
// @Failure      400  {object}  map[string]interface{} "登录失败信息"
// @Failure      500  {object}  map[string]interface{} "服务器错误信息"
// @Router       /admin/login [post]
// @Security     None
func (c *SuperAdminController) Login() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		log.Warn("超级管理员登录表单无效 | %v | 错误: %v", requestID, err)
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	form.Email = strings.TrimSpace(form.Email)
	form.Password = strings.TrimSpace(form.Password)

	log.Info("超级管理员尝试登录 | %v | 邮箱: %s", requestID, form.Email)

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).CheckPassword(form.Email, form.Password); err != nil {
		log.Warn("超级管理员登录失败 | %v | 邮箱: %s | 错误: %v", requestID, form.Email, err)
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "login failed",
		})
		return
	}

	// Get admin info
	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminByEmail(form.Email)
	if err != nil {
		log.Error("获取超级管理员信息失败 | %v | 邮箱: %s | 错误: %v", requestID, form.Email, err)
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to get admin info",
		})
		return
	}

	// Generate token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateSuperAdminToken(admin)
	if err != nil {
		log.Error("生成超级管理员令牌失败 | %v | 管理员ID: %d | 错误: %v", requestID, admin.ID, err)
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to generate token",
		})
		return
	}

	log.Info("超级管理员登录成功 | %v | 管理员ID: %d", requestID, admin.ID)
	c.Ctx.JSON(200, gin.H{
		"message": "login success",
		"token":   token,
	})
}

// 2.GetSuperAdmins 获取超级管理员列表
// @Summary      获取超级管理员列表
// @Description  分页获取所有超级管理员用户列表
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        search query string false "搜索关键词(邮箱等)"
// @Param        pageSize query int false "每页条数, 默认为10"
// @Param        pageNum query int false "页码, 默认为1"
// @Param        desc query bool false "是否降序排序, 默认为false"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /admin/super_admin [get]
// @Security     BearerAuth
func (c *SuperAdminController) GetSuperAdmins() {
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

	admins, paginationResult, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdmins(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       admins,
		"pagination": paginationResult,
	})
}

// 3.CreateSuperAdmin 创建超级管理员
// @Summary      创建超级管理员
// @Description  创建新的超级管理员账号
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        request body SuperAdminCreateRequest true "超级管理员信息"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /admin/super_admin [post]
// @Security     BearerAuth
func (c *SuperAdminController) CreateSuperAdmin() {
	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	superAdmin := &base_models.SuperAdmin{
		Email:    strings.TrimSpace(form.Email),
		Password: strings.TrimSpace(form.Password),
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).CreateSuperAdmin(superAdmin); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create super admin failed",
		})
		return
	}

	superAdmin.Password = ""
	c.Ctx.JSON(200, gin.H{
		"message": "create super admin success",
		"data":    superAdmin,
	})
}

// 4.DeleteSuperAdmin 删除超级管理员
// @Summary      删除超级管理员
// @Description  根据ID批量删除超级管理员账号(不能删除自己)
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        request body SuperAdminDeleteRequest true "要删除的超级管理员ID列表"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /admin/super_admin [delete]
// @Security     BearerAuth
func (c *SuperAdminController) DeleteSuperAdmin() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid token claims format"})
		return
	}

	currentIDFloat, ok := mapClaims["id"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid id in token"})
		return
	}
	currentID := uint(currentIDFloat)

	for _, id := range form.IDs {
		if id == currentID {
			c.Ctx.JSON(400, gin.H{
				"error": "cannot delete yourself",
				"id":    id,
			})
			return
		}
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).DeleteSuperAdmins(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete success"})
}

// 5.ResetPassword 重置超级管理员密码
// @Summary      重置超级管理员密码
// @Description  根据ID重置指定超级管理员的密码
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        request body SuperAdminResetPasswordRequest true "重置密码信息"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /admin/super_admin/reset_password [post]
// @Security     BearerAuth
func (c *SuperAdminController) ResetPassword() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid token claims format"})
		return
	}

	currentIDFloat, ok := mapClaims["id"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid id in token"})
		return
	}
	currentID := uint(currentIDFloat)

	if form.ID == currentID {
		c.Ctx.JSON(400, gin.H{
			"error": "cannot reset your own password, use change password instead",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error": "failed to hash password",
		})
		return
	}

	// 获取管理员对象
	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(form.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashedPassword),
	}); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "reset password success",
	})
}

// 6.ChangePassword 修改自己的密码
// @Summary      修改自己的密码
// @Description  超级管理员修改自己的密码
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        request body SuperAdminChangePasswordRequest true "密码修改信息"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /admin/super_admin/update_password [post]
// @Security     BearerAuth
func (c *SuperAdminController) ChangePassword() {
	var form struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid token claims format"})
		return
	}

	currentIDFloat, ok := mapClaims["id"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid id in token"})
		return
	}
	currentID := uint(currentIDFloat)

	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(currentID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(form.OldPassword)); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": "old password is incorrect",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error": "failed to hash password",
		})
		return
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashedPassword),
	}); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "change password success",
	})
}

// 7.GetOne 获取单个超级管理员信息
// @Summary      获取单个超级管理员信息
// @Description  根据ID获取超级管理员详细信息
// @Tags         SuperAdmin
// @Accept       json
// @Produce      json
// @Param        id path int true "超级管理员ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /admin/super_admin/{id} [get]
// @Security     BearerAuth
func (c *SuperAdminController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": "invalid id",
		})
		return
	}

	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	admin.Password = ""
	c.Ctx.JSON(200, gin.H{
		"data": admin,
	})
}
