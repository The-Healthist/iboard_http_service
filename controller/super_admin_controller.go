package http_controller

import (
	"fmt"
	"strings"

	"github.com/The-Healthist/iboard_http_service/models"
	"github.com/The-Healthist/iboard_http_service/services"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type InterfaceSuperAdminController interface {
	Login()
	GetSuperAdmins()
	CreateSuperAdmin()
	DeleteSuperAdmin()
	ResetPassword()
	ChangePassword()
}

type SuperAdminController struct {
	ctx          *gin.Context
	service      services.InterfaceSuperAdminService
	jwtService   *services.IJWTService
	emailService *services.IEmailService
}

func NewSuperAdminController(
	ctx *gin.Context,
	service services.InterfaceSuperAdminService,
	jwtService *services.IJWTService,
	emailService *services.IEmailService,
) InterfaceSuperAdminController {
	return &SuperAdminController{
		ctx:          ctx,
		service:      service,
		jwtService:   jwtService,
		emailService: emailService,
	}
}

// 1.login
func (c *SuperAdminController) Login() {
	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	form.Email = strings.TrimSpace(form.Email)
	form.Password = strings.TrimSpace(form.Password)

	if err := c.service.CheckPassword(form.Email, form.Password); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "login failed",
		})
		return
	}

	if c.jwtService == nil {
		c.ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "login success",
		"token":   (*c.jwtService).GenerateToken(form.Email, true),
	})
}

// 2.get super admins
func (c *SuperAdminController) GetSuperAdmins() {
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

	admins, paginationResult, err := c.service.GetSuperAdmins(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       admins,
		"pagination": paginationResult,
	})
}

// 3.create super admin
func (c *SuperAdminController) CreateSuperAdmin() {
	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	superAdmin := &models.SuperAdmin{
		Email:    strings.TrimSpace(form.Email),
		Password: strings.TrimSpace(form.Password),
	}

	if err := c.service.CreateSuperAdmin(superAdmin); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create super admin failed",
		})
		return
	}

	superAdmin.Password = ""
	c.ctx.JSON(200, gin.H{
		"message": "create super admin success",
		"data":    superAdmin,
	})
}

// 4.delete super admin
func (c *SuperAdminController) DeleteSuperAdmin() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.DeleteSuperAdmins(form.IDs); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete super admin success"})
}

func (c *SuperAdminController) ResetPassword() {
	var form struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	admin, err := c.service.GetSuperAdminById(form.ID)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	randPassword := utils.RandStr(8, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(randPassword), bcrypt.DefaultCost)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashPassword),
	}); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if c.emailService != nil {
		emailContent := fmt.Sprintf("Your password has been reset to: %s", randPassword)
		if err := (*c.emailService).SendEmail(
			[]string{admin.Email},
			"Password Reset Notification",
			emailContent,
		); err != nil {
			c.ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	c.ctx.JSON(200, gin.H{"message": "reset password success"})
}

// 5.change password
func (c *SuperAdminController) ChangePassword() {
	var form struct {
		ID          uint   `json:"id" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	admin, err := c.service.GetSuperAdminById(form.ID)
	if err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   "admin not found",
			"message": err.Error(),
		})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "password encryption failed",
		})
		return
	}

	if err := c.service.UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashPassword),
	}); err != nil {
		c.ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "update password failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "change password success"})
}
