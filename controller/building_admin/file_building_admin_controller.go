package building_admin_controllers

import (
	"net/http"
	"path/filepath"
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/utils/response"
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

	// 构建查询参数
	query := make(map[string]interface{})
	if mimeType := c.ctx.Query("mimeType"); mimeType != "" {
		query["mimeType"] = mimeType
	}

	// 分页参数
	paginate := map[string]interface{}{
		"pageSize": 10,
		"pageNum":  1,
		"desc":     true,
	}

	files, pagination, err := c.service.Get(email, query, paginate)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": pagination,
	})
}

func (c *BuildingAdminFileController) GetFile() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": file})
}

func (c *BuildingAdminFileController) UploadFile() {
	// 获取文件
	file, err := c.ctx.FormFile("file")
	if err != nil {
		response.Error(c.ctx, http.StatusBadRequest, "No file uploaded")
		return
	}

	// 验证文件大小
	if file.Size > 10<<20 { // 10MB
		response.Error(c.ctx, http.StatusBadRequest, "File too large")
		return
	}

	// 验证文件类型
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"application/pdf": true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		response.Error(c.ctx, http.StatusBadRequest, "Invalid file type")
		return
	}

	email := c.ctx.GetString("email")

	filename := filepath.Base(file.Filename)
	dst := filepath.Join("uploads", filename)

	if err := c.ctx.SaveUploadedFile(file, dst); err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	fileRecord := &base_models.File{
		Path:         dst,
		Size:         file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		Uploader:     email,
		UploaderType: "building_admin",
	}

	if err := c.service.Create(fileRecord, email); err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to create file record"})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "File uploaded successfully",
		"data":    fileRecord,
	})
}

func (c *BuildingAdminFileController) UpdateFile() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ctx.ShouldBindJSON(&updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File updated successfully"})
}

func (c *BuildingAdminFileController) DeleteFile() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "File deleted successfully"})
}

func (c *BuildingAdminFileController) DownloadFile() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid file ID"})
		return
	}

	file, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.File(file.Path)
}
