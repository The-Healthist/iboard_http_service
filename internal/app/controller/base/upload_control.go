package http_base_controller

import (
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type UploadController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewUploadController(ctx *gin.Context, container *container.ServiceContainer) *UploadController {
	return &UploadController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncUpload returns a gin.HandlerFunc for the specified method
func HandleFuncUpload(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "getUploadParams":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.GetUploadParams()
		}
	case "uploadCallback":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.UploadCallback()
		}
	case "uploadCallbackSync":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.UploadCallbackSync()
		}
	case "getUploadParamsSync":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.GetUploadParamsSync()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *UploadController) GetUploadParams() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	log.Info("处理上传参数请求 | %v", requestID)

	// 默认上传者信息（使用默认超级管理员）
	var uploaderID uint = 1
	var uploaderType string = "superAdmin"
	var uploaderEmail string = "admin@example.com"

	// Try to get JWT claims if available (optional now)
	claims, exists := c.Ctx.Get("claims")
	if exists {
		mapClaims, ok := claims.(jwt.MapClaims)
		if ok {
			// Parse and validate claims if JWT is present
			isAdmin, _ := mapClaims["isAdmin"].(bool)
			isBuildingAdmin, _ := mapClaims["isBuildingAdmin"].(bool)

			if isAdmin {
				// Handle super admin case
				if id, ok := mapClaims["id"].(float64); ok {
					uploaderID = uint(id)
					uploaderType = "superAdmin"
					if email, ok := mapClaims["email"].(string); ok {
						uploaderEmail = email
					}
					log.Info("识别为超级管理员 | %v | ID: %d, 邮箱: %s", requestID, uploaderID, uploaderEmail)
				}
			} else if isBuildingAdmin {
				// Handle building admin case
				if id, ok := mapClaims["id"].(float64); ok {
					uploaderID = uint(id)
					uploaderType = "buildingAdmin"
					if email, ok := mapClaims["email"].(string); ok {
						uploaderEmail = email
					}
					log.Info("识别为建筑管理员 | %v | ID: %d, 邮箱: %s", requestID, uploaderID, uploaderEmail)
				}
			}
		}
	} else {

	}

	// Save uploader info to cache
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveUploaderInfo(uploaderID, uploaderType, uploaderEmail); err != nil {
		log.Error("保存上传者信息失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process uploader info",
		})
		return
	}

	// Process file upload parameters
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("无效的请求参数 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Generate directory and filename based on file type
	ext := path.Ext(req.FileName)
	newFileName := uuid.New().String() + ext

	// Determine directory based on file type
	var dir string
	switch strings.ToLower(ext) {
	case ".apk":
		dir = "iboard/apks/"
	case ".pdf":
		dir = "iboard/pdf/"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg":
		dir = "iboard/image/"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v":
		dir = "iboard/video/"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma":
		dir = "iboard/audio/"
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		dir = "iboard/document/"
	default:
		dir = "iboard/other/"
	}

	fullPath := dir + newFileName

	// Get upload policy - use callback for all files including APK
	var policy map[string]interface{}
	var err error

	// Use normal callback method for all files (including APK)
	policy, err = c.Container.GetService("upload").(base_services.IUploadService).GetUploadParams(fullPath)
	log.Info("使用回调上传方式处理文件 | %v | 文件类型: %s", requestID, path.Ext(req.FileName))

	if err != nil {
		log.Error("获取上传参数失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save filename mapping
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveFileNameMapping(newFileName, fullPath); err != nil {
		log.Error("保存文件名映射失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save filename mapping",
		})
		return
	}

	log.Info("成功生成文件上传参数 | %v | 文件: %s", requestID, fullPath)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"data": policy,
	})
}

func (c *UploadController) UploadCallback() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	log.Info("接收到上传回调请求 | %v", requestID)

	// Get content-md5 from header
	contentMD5 := c.Ctx.GetHeader("content-md5")
	log.Debug("Content-MD5头信息 | %v | %s", requestID, contentMD5)

	// Read and log raw request body
	body, err := io.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		log.Error("读取请求体失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}
	log.Debug("原始请求体 | %v | %s", requestID, string(body))

	// Restore request body for form parsing
	c.Ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Parse form data
	if err := c.Ctx.Request.ParseForm(); err != nil {
		log.Error("解析表单失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	log.Debug("表单值 | %v | %+v", requestID, c.Ctx.Request.Form)

	// Get callback data
	objectPath := c.Ctx.Request.Form.Get("object")
	sizeStr := c.Ctx.Request.Form.Get("size")
	mimeType := c.Ctx.Request.Form.Get("mimeType")
	heightStr := c.Ctx.Request.Form.Get("height")
	widthStr := c.Ctx.Request.Form.Get("width")

	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	height, _ := strconv.Atoi(heightStr)
	width, _ := strconv.Atoi(widthStr)

	log.Debug("解析的回调数据 | %v | object=%s, size=%d, mimeType=%s, height=%d, width=%d",
		requestID, objectPath, size, mimeType, height, width)

	// Get stored dirPath from Redis
	dirPath, err := c.Container.GetService("upload").(base_services.IUploadService).GetLatestDirPath()
	if err != nil {
		log.Error("从Redis获取存储的dirPath失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stored dirPath",
		})
		return
	}
	log.Debug("从Redis检索到存储的dirPath | %v | %s", requestID, dirPath)

	// Compare paths
	if objectPath != dirPath {
		log.Warn("路径不匹配 | %v | 回调=%s, 存储=%s", requestID, objectPath, dirPath)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Path mismatch",
		})
		return
	}
	log.Debug("路径验证成功 | %v", requestID)

	// Get stored uploader info from cache
	uploaderID, uploaderType, uploaderEmail, err := c.Container.GetService("upload").(base_services.IUploadService).GetLatestUploaderInfo()
	if err != nil {
		log.Error("从缓存获取上传者信息失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get uploader info",
		})
		return
	}
	log.Debug("检索到上传者信息 | %v | ID=%d, 类型=%s, 邮箱=%s", requestID, uploaderID, uploaderType, uploaderEmail)

	// Create file record
	var fileUploaderType field.FileUploaderType
	if uploaderType == "superAdmin" {
		fileUploaderType = field.UploaderTypeSuperAdmin
	} else {
		fileUploaderType = field.UploaderTypeBuildingAdmin
	}

	// Construct full path
	fullPath := os.Getenv("HOST") + "/" + objectPath
	log.Debug("构建完整文件路径 | %v | %s", requestID, fullPath)

	file := &base_models.File{
		Path:         fullPath,
		Size:         size,
		MimeType:     mimeType,
		Md5:          contentMD5,
		Oss:          "aws",
		Uploader:     uploaderEmail,
		UploaderType: fileUploaderType,
		UploaderID:   uploaderID,
	}

	log.Debug("创建文件记录 | %v | 详情: %+v", requestID, file)

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
		log.Error("创建文件记录失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create file record",
		})
		return
	}
	log.Info("文件记录创建成功 | %v | 文件ID: %d", requestID, file.ID)

	// Return success response
	c.Ctx.JSON(http.StatusOK, gin.H{
		"Status": "OK",
	})
}
func (c *UploadController) UploadCallbackSync() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	log.Info("接收到同步上传回调请求 | %v", requestID)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"Status": "OK",
	})
}
func (c *UploadController) GetUploadParamsSync() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	log.Info("处理同步上传参数请求 | %v", requestID)

	// 默认上传者信息（使用默认超级管理员，特别是同步服务）
	var uploaderID uint = 1
	var uploaderType string = "superAdmin"
	var uploaderEmail string = "admin@example.com"

	// Try to get JWT claims if available (optional now)
	claims, exists := c.Ctx.Get("claims")
	if exists {
		mapClaims, ok := claims.(jwt.MapClaims)
		if ok {
			// Parse and validate claims if JWT is present
			isAdmin, _ := mapClaims["isAdmin"].(bool)
			isBuildingAdmin, _ := mapClaims["isBuildingAdmin"].(bool)

			if isAdmin {
				// Handle super admin case
				if id, ok := mapClaims["id"].(float64); ok {
					uploaderID = uint(id)
					uploaderType = "superAdmin"
					if email, ok := mapClaims["email"].(string); ok {
						uploaderEmail = email
					}
					log.Info("识别为超级管理员 | %v | ID: %d, 邮箱: %s", requestID, uploaderID, uploaderEmail)
				}
			} else if isBuildingAdmin {
				// Handle building admin case
				if id, ok := mapClaims["id"].(float64); ok {
					uploaderID = uint(id)
					uploaderType = "buildingAdmin"
					if email, ok := mapClaims["email"].(string); ok {
						uploaderEmail = email
					}
					log.Info("识别为建筑管理员 | %v | ID: %d, 邮箱: %s", requestID, uploaderID, uploaderEmail)
				}
			}
		}
	} else {
	}

	// Save uploader info to cache
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveUploaderInfo(uploaderID, uploaderType, uploaderEmail); err != nil {
		log.Error("保存上传者信息失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process uploader info",
		})
		return
	}

	// Process file upload parameters
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		log.Warn("无效的请求参数 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Generate directory and filename based on file type
	ext := path.Ext(req.FileName)
	newFileName := uuid.New().String() + ext

	// Determine directory based on file type
	var dir string
	switch strings.ToLower(ext) {
	case ".apk":
		dir = "iboard/apks/"
	case ".pdf":
		dir = "iboard/pdf/"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg":
		dir = "iboard/image/"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v":
		dir = "iboard/video/"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma":
		dir = "iboard/audio/"
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt":
		dir = "iboard/document/"
	default:
		dir = "iboard/other/"
	}

	fullPath := dir + newFileName

	// Get upload policy
	policy, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParamsSync(fullPath)
	if err != nil {
		log.Error("获取上传参数失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save filename mapping
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveFileNameMapping(newFileName, fullPath); err != nil {
		log.Error("保存文件名映射失败 | %v | %v", requestID, err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save filename mapping",
		})
		return
	}

	log.Info("成功生成文件上传参数 | %v | 文件: %s", requestID, fullPath)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"data": policy,
	})
}
