package building_admin_controllers

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type BuildingAdminFileController struct {
	ctx     *gin.Context
	service building_admin_services.InterfaceBuildingAdminFileService
}

func NewBuildingAdminFileController(ctx *gin.Context, service building_admin_services.InterfaceBuildingAdminFileService) *BuildingAdminFileController {
	return &BuildingAdminFileController{
		ctx:     ctx,
		service: service,
	}
}

func (c *BuildingAdminFileController) GetFiles() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var searchQuery struct {
		Search   string `form:"search"`
		MimeType string `form:"mimeType"`
		Oss      string `form:"oss"`
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

	files, paginationResult, err := c.service.Get(email, queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminFileController) GetFile() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Get file success",
		"data":    file,
	})
}

func (c *BuildingAdminFileController) UploadFile() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var form struct {
		Path     string                 `json:"path" binding:"required"`
		Size     int64                  `json:"size" binding:"required"`
		Md5      string                 `json:"md5" binding:"required"`
		MimeType string                 `json:"mimeType" binding:"required"`
		Oss      string                 `json:"oss"`
		Type     field.FileUploaderType `json:"type"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// 计算文件的 MD5
	hasher := md5.New()
	hasher.Write([]byte(form.Path))
	calculatedMd5 := hex.EncodeToString(hasher.Sum(nil))

	// 验证 MD5
	if calculatedMd5 != form.Md5 {
		c.ctx.JSON(400, gin.H{
			"error":   "MD5 mismatch",
			"message": "File integrity check failed",
		})
		return
	}

	file := &base_models.File{
		Path:         form.Path,
		Size:         form.Size,
		Md5:          form.Md5,
		MimeType:     form.MimeType,
		Oss:          form.Oss,
		Uploader:     email,
		UploaderType: "building_admin",
	}

	if err := c.service.Create(file, email); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create file failed",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create file success",
		"data":    file,
	})
}

func (c *BuildingAdminFileController) UpdateFile() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	var form struct {
		Path     string `json:"path"`
		MimeType string `json:"mimeType"`
		Oss      string `json:"oss"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Path != "" {
		updates["path"] = form.Path
	}
	if form.MimeType != "" {
		updates["mime_type"] = form.MimeType
	}
	if form.Oss != "" {
		updates["oss"] = form.Oss
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取更新后的文件信息
	updatedFile, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(200, gin.H{
			"message": "update file success, but failed to load updated data",
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "update file success",
		"data":    updatedFile,
	})
}

func (c *BuildingAdminFileController) DeleteFile() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete file success"})
}

func (c *BuildingAdminFileController) DownloadFile() {
	email := c.ctx.GetString("email")
	if email == "" {
		c.ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.ctx.File(file.Path)
}
