package http_base_controller

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceFileController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

type FileController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewFileController(ctx *gin.Context, container *container.ServiceContainer) *FileController {
	return &FileController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncFile returns a gin.HandlerFunc for the specified method
func HandleFuncFile(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewFileController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// 1.Create 创建文件
// @Summary      创建文件
// @Description  创建单个文件记录
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        file body object true "文件信息"
// @Param        path formData string true "文件路径" example:"/uploads/documents/report.pdf"
// @Param        size formData int64 true "文件大小" example:"1024"
// @Param        mimeType formData string true "文件类型" example:"application/pdf"
// @Param        oss formData string true "对象存储服务" example:"aws"
// @Param        uploaderType formData string true "上传者类型" example:"super_admin"
// @Param        uploaderId formData uint true "上传者ID" example:"1"
// @Param        md5 formData string false "文件MD5" example:"a1b2c3d4e5f6g7h8i9j0"
// @Success      200  {object}  map[string]interface{} "返回创建的文件信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/file [post]
// @Security     BearerAuth
func (c *FileController) Create() {
	var form struct {
		Path         string                 `json:"path" binding:"required" example:"/uploads/documents/report.pdf"`
		Size         int64                  `json:"size" binding:"required" example:"1024"`
		MimeType     string                 `json:"mimeType" binding:"required" example:"application/pdf"`
		Oss          string                 `json:"oss" binding:"required" example:"aws"`
		UploaderType field.FileUploaderType `json:"uploaderType" binding:"required" example:"super_admin"`
		UploaderID   uint                   `json:"uploaderId" binding:"required" example:"1"`
		Md5          string                 `json:"md5" example:"a1b2c3d4e5f6g7h8i9j0"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	if form.Md5 == "" {
		data := form.Path + time.Now().String()
		hash := md5.Sum([]byte(data))
		form.Md5 = hex.EncodeToString(hash[:])
	}

	file := &base_models.File{
		Path:         form.Path,
		Size:         form.Size,
		MimeType:     form.MimeType,
		Oss:          form.Oss,
		UploaderType: form.UploaderType,
		UploaderID:   form.UploaderID,
		Md5:          form.Md5,
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create file failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create file success",
		"data":    file,
	})
}

// 2.CreateMany 批量创建文件
// @Summary      批量创建文件
// @Description  批量创建多个文件记录
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        files body array true "文件信息数组"
// @Param        path formData string true "文件路径" example:"/uploads/documents/report.pdf"
// @Param        size formData int64 true "文件大小" example:"1024"
// @Param        mimeType formData string true "文件类型" example:"application/pdf"
// @Param        oss formData string true "对象存储服务" example:"aws"
// @Param        uploaderType formData string true "上传者类型" example:"super_admin"
// @Param        uploaderId formData uint true "上传者ID" example:"1"
// @Param        md5 formData string false "文件MD5" example:"a1b2c3d4e5f6g7h8i9j0"
// @Success      200  {object}  map[string]interface{} "返回创建的文件信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/files [post]
// @Security     BearerAuth
func (c *FileController) CreateMany() {
	var forms []struct {
		Path         string                 `json:"path" binding:"required" example:"/uploads/documents/report.pdf"`
		Size         int64                  `json:"size" binding:"required" example:"1024"`
		MimeType     string                 `json:"mimeType" binding:"required" example:"application/pdf"`
		Oss          string                 `json:"oss" binding:"required" example:"aws"`
		UploaderType field.FileUploaderType `json:"uploaderType" binding:"required" example:"super_admin"`
		UploaderID   uint                   `json:"uploaderId" binding:"required" example:"1"`
		Md5          string                 `json:"md5" example:"a1b2c3d4e5f6g7h8i9j0"`
	}

	if err := c.Ctx.ShouldBindJSON(&forms); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var files []*base_models.File
	for _, form := range forms {
		if form.Md5 == "" {
			data := form.Path + time.Now().String()
			hash := md5.Sum([]byte(data))
			form.Md5 = hex.EncodeToString(hash[:])
		}

		file := &base_models.File{
			Path:         form.Path,
			Size:         form.Size,
			MimeType:     form.MimeType,
			Oss:          form.Oss,
			UploaderType: form.UploaderType,
			UploaderID:   form.UploaderID,
			Md5:          form.Md5,
		}
		files = append(files, file)
	}

	for _, file := range files {
		if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "create file failed",
				"file":    file,
			})
			return
		}
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create files success",
		"data":    files,
	})
}

// 3.Get 获取文件列表
// @Summary      获取文件列表
// @Description  根据查询条件获取文件列表
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        mimeType query string false "文件类型" example:"application/pdf"
// @Param        oss query string false "对象存储服务" example:"aws"
// @Param        uploaderType query string false "上传者类型" example:"super_admin"
// @Param        pageSize query int false "每页数量" default(10)
// @Param        pageNum query int false "页码" default(1)
// @Param        desc query bool false "是否降序" default(false)
// @Success      200  {object}  map[string]interface{} "返回文件列表和分页信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/file [get]
// @Security     BearerAuth
func (c *FileController) Get() {
	var searchQuery struct {
		MimeType     string `form:"mimeType" example:"application/pdf"`
		Oss          string `form:"oss" example:"aws"`
		UploaderType string `form:"uploaderType" example:"super_admin"`
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

	files, paginationResult, err := c.Container.GetService("file").(base_services.InterfaceFileService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       files,
		"pagination": paginationResult,
	})
}

// 4.Update 更新文件
// @Summary      更新文件
// @Description  更新文件信息
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        file body object true "文件更新信息"
// @Param        id formData uint true "文件ID" example:"1"
// @Param        path formData string false "文件路径" example:"/uploads/documents/updated_report.pdf"
// @Param        size formData int64 false "文件大小" example:"2048"
// @Param        mimeType formData string false "文件类型" example:"application/pdf"
// @Param        oss formData string false "对象存储服务" example:"aws"
// @Param        uploader formData string false "上传者" example:"admin@example.com"
// @Param        uploaderType formData string false "上传者类型" example:"super_admin"
// @Success      200  {object}  map[string]interface{} "更新成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/file [put]
// @Security     BearerAuth
func (c *FileController) Update() {
	var form struct {
		ID           uint   `json:"id" binding:"required" example:"1"`
		Path         string `json:"path" example:"/uploads/documents/updated_report.pdf"`
		Size         int64  `json:"size" example:"2048"`
		MimeType     string `json:"mimeType" example:"application/pdf"`
		Oss          string `json:"oss" example:"aws"`
		Uploader     string `json:"uploader" example:"admin@example.com"`
		UploaderType string `json:"uploaderType" example:"super_admin"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Path != "" {
		updates["path"] = form.Path
	}
	if form.Size != 0 {
		updates["size"] = form.Size
	}
	if form.MimeType != "" {
		updates["mime_type"] = form.MimeType
	}
	if form.Oss != "" {
		updates["oss"] = form.Oss
	}
	if form.Uploader != "" {
		updates["uploader"] = form.Uploader
	}
	if form.UploaderType != "" {
		updates["uploader_type"] = form.UploaderType
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update file success"})
}

// 5.Delete 删除文件
// @Summary      删除文件
// @Description  删除一个或多个文件
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        ids body object true "文件ID列表"
// @Param        ids.ids body []uint true "文件ID数组" example:"[1,2,3]"
// @Success      200  {object}  map[string]interface{} "删除成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/file [delete]
// @Security     BearerAuth
func (c *FileController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required" example:"[1,2,3]"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete file success"})
}

// 6.GetOne 获取单个文件
// @Summary      获取单个文件
// @Description  根据ID获取文件详细信息
// @Tags         File
// @Accept       json
// @Produce      json
// @Param        id path int true "文件ID" example:"1"
// @Success      200  {object}  map[string]interface{} "返回文件详细信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "文件不存在"
// @Router       /admin/file/{id} [get]
// @Security     BearerAuth
func (c *FileController) GetOne() {
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
		c.Ctx.JSON(400, gin.H{"error": "invalid file ID"})
		return
	}

	file, err := c.Container.GetService("file").(base_services.InterfaceFileService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get file success",
		"data":    file,
	})
}
