package http_base_controller

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	SyncNotice()
}

type BuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingController(ctx *gin.Context, container *container.ServiceContainer) *BuildingController {
	return &BuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetOne()
		}
	case "login":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Login()
		}
	case "getBuildingAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetBuildingAdvertisements()
		}
	case "getBuildingNotices":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetBuildingNotices()
		}
	case "syncNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.SyncNotice()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Create creates a new building
func (c *BuildingController) Create() {
	var form struct {
		Name     string `json:"name" binding:"required"`
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
		Remark   string `json:"remark"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
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

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Create(building); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create building success",
		"data":    building,
	})
}

func (c *BuildingController) Get() {
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

	buildings, paginationResult, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       buildings,
		"pagination": paginationResult,
	})
}

func (c *BuildingController) Update() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Name     string `json:"name"`
		IsmartID string `json:"ismartId"`
		Password string `json:"password"`
		Remark   string `json:"remark"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
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

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update building success"})
}

func (c *BuildingController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Get buildings to delete
	var buildings []base_models.Building
	if err := tx.Where("id IN ?", form.IDs).Find(&buildings).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get buildings",
			"message": err.Error(),
		})
		return
	}

	// 2. Remove advertisement associations
	if err := tx.Exec("DELETE FROM advertisement_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind advertisements",
			"message": err.Error(),
		})
		return
	}

	// 3. Remove notice associations
	if err := tx.Exec("DELETE FROM notice_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind notices",
			"message": err.Error(),
		})
		return
	}

	// 4. Remove admin associations
	if err := tx.Exec("DELETE FROM building_admins_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind admins",
			"message": err.Error(),
		})
		return
	}

	// 5. Delete buildings
	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Delete(form.IDs); err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete buildings",
			"message": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete building success"})
}

func (c *BuildingController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building ID",
			"message": "Please check the ID format",
		})
		return
	}

	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get building",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get building success",
		"data":    building,
	})
}

func (c *BuildingController) Login() {
	var form struct {
		IsmartID string `json:"ismartId" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByCredentials(form.IsmartID, form.Password)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateBuildingToken(building)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to generate token",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
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

func (c *BuildingController) GetBuildingAdvertisements() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	advertisements, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetBuildingAdvertisements(buildingId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get advertisements",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get advertisements success",
		"data":    advertisements,
	})
}

func (c *BuildingController) GetBuildingNotices() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid claims format"})
		return
	}

	buildingIdFloat, ok := claimsMap["buildingId"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid building id format"})
		return
	}

	buildingId := uint(buildingIdFloat)
	notices, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetBuildingNotices(buildingId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get notices",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get notices success",
		"data":    notices,
	})
}

func (c *BuildingController) SyncNotice() {
	// 从URL参数获取建筑ID
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "无效的建筑ID",
			"message": "请检查ID格式",
		})
		return
	}

	// 获取建筑信息以获取ismartId
	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "获取建筑信息失败",
			"message": err.Error(),
		})
		return
	}

	// 准备请求老系统的API
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "准备请求数据失败",
			"message": err.Error(),
		})
		return
	}

	// 请求老系统获取通知列表
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "从老系统获取通知失败",
			"message": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 解析响应数据
	var respBody []struct {
		ID        int    `json:"id"`
		MessTitle string `json:"mess_title"`
		MessType  string `json:"mess_type"`
		MessFile  string `json:"mess_file"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "解析响应数据失败",
			"message": err.Error(),
		})
		return
	}

	// 处理每个通知
	var successCount int
	var failedNotices []string
	var hasSyncedCount int

	for _, oldNotice := range respBody {
		// 下载PDF文件
		fileResp, err := http.Get(oldNotice.MessFile)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("下载通知ID %d 的文件失败: %v", oldNotice.ID, err))
			continue
		}

		// 读取文件内容
		fileContent, err := io.ReadAll(fileResp.Body)
		fileResp.Body.Close()
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("读取通知ID %d 的文件内容失败: %v", oldNotice.ID, err))
			continue
		}
		//计算文件大小
		fileSize := len(fileContent)
		// 计算文件md5
		md5Hash := md5.Sum(fileContent)
		md5Str := hex.EncodeToString(md5Hash[:])
		//获取文件mimeType
		mimeType := http.DetectContentType(fileContent)

		log.Println(mimeType, fileSize, md5Str) //////////////////////////////////////

		// 获取上传者信息从token
		claims, exists := c.Ctx.Get("claims")
		if !exists {
			failedNotices = append(failedNotices, fmt.Sprintf("获取通知ID %d 的上传者信息失败: token claims not found", oldNotice.ID))
			continue
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			failedNotices = append(failedNotices, fmt.Sprintf("获取通知ID %d 的上传者信息失败: invalid claims format", oldNotice.ID))
			continue
		}

		var uploaderID uint
		var uploaderType field.FileUploaderType
		var uploaderEmail string

		if isAdmin, ok := mapClaims["isAdmin"].(bool); ok && isAdmin {
			if id, ok := mapClaims["id"].(float64); ok {
				uploaderID = uint(id)
				uploaderType = field.UploaderTypeSuperAdmin
				if email, ok := mapClaims["email"].(string); ok {
					uploaderEmail = email
				}
			}
		}

		if uploaderID == 0 {
			failedNotices = append(failedNotices, fmt.Sprintf("获取通知ID %d 的上传者信息失败: invalid uploader info", oldNotice.ID))
			continue
		}

		// 生成文件名并获取上传参数
		currentTime := time.Now()
		dir := currentTime.Format("2006-01-02") + "/"
		fileName := fmt.Sprintf("%s.pdf", uuid.New().String())
		objectKey := dir + fileName
		uploadParams, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParamsSync(objectKey)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("获取通知ID %d 的上传参数失败: %v", oldNotice.ID, err))
			continue
		}

		// 创建文件记录前，先检查MD5是否存在
		var existingFile base_models.File
		var shouldUpload bool = true
		var shouldAddFile bool = true
		var fileForNotice *base_models.File

		err = databases.DB_CONN.Where("md5 = ?", md5Str).First(&existingFile).Error
		if err == nil {
			// 文件已存在，检查是否已经与通知关联
			var notice base_models.Notice
			var noticeExists bool

			// 检查文件是否与通知关联
			err := databases.DB_CONN.Where("file_id = ?", existingFile.ID).First(&notice).Error
			noticeExists = err == nil

			if noticeExists {
				// 检查通知是否与当前建筑关联
				var count int64
				err = databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ? AND building_id = ?", notice.ID, id).
					Count(&count).Error
				if err == nil && count > 0 {
					// 通知已经与当前建筑关联，跳过处理
					hasSyncedCount++
					log.Printf("通知ID %d 已经同步过，跳过处理", oldNotice.ID)
					continue
				}

				if err := databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ?", notice.ID).
					Count(&count).Error; err == nil && count > 0 {
					// 通知已绑定其他建筑，需要创建新文件和通知
					shouldUpload = true
					shouldAddFile = true
					fileForNotice = &base_models.File{
						Path:         uploadParams["host"].(string) + "/" + objectKey,
						Size:         int64(fileSize),
						MimeType:     mimeType,
						Oss:          "aws",
						UploaderType: uploaderType,
						UploaderID:   uploaderID,
						Uploader:     uploaderEmail,
						Md5:          md5Str,
					}
				} else {
					// 通知存在但未绑定任何建筑，只需绑定到当前建筑
					shouldUpload = false
					shouldAddFile = false
					fileForNotice = &existingFile

					// 直接绑定通知到当前建筑
					if err := databases.DB_CONN.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)", notice.ID, id).Error; err != nil {
						failedNotices = append(failedNotices, fmt.Sprintf("绑定通知ID %d 到建筑物失败: %v", oldNotice.ID, err))
						continue
					}
					successCount++
					continue
				}
			} else {
				// 文件存在但未关联通知，使用现有文件创建新通知
				shouldUpload = false
				shouldAddFile = false
				fileForNotice = &existingFile
			}
		} else if err == gorm.ErrRecordNotFound {
			// 文件不存在，需要全部创建
			shouldUpload = true
			shouldAddFile = true
			fileForNotice = &base_models.File{
				Path:         uploadParams["host"].(string) + "/" + objectKey,
				Size:         int64(fileSize),
				MimeType:     mimeType,
				Oss:          "aws",
				UploaderType: uploaderType,
				UploaderID:   uploaderID,
				Uploader:     uploaderEmail,
				Md5:          md5Str,
			}
		} else {
			failedNotices = append(failedNotices, fmt.Sprintf("检查通知ID %d 的文件MD5失败: %v", oldNotice.ID, err))
			continue
		}

		// 如果需要创建新文件
		if shouldAddFile {
			if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(fileForNotice); err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("创建通知ID %d 的文件记录失败: %v", oldNotice.ID, err))
				continue
			}
		}

		// 如果需要上传文件
		if shouldUpload {
			// 准备上传表单数据
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			// 按照前端的完全相同顺序添加表单字段
			formFields := []struct {
				key      string
				paramKey string
			}{
				{key: "key", paramKey: "dir"},
				{key: "policy", paramKey: "policy"},
				{key: "OSSAccessKeyId", paramKey: "accessid"},
				{key: "success_action_status", paramKey: ""},
				{key: "callback", paramKey: "callback"},
				{key: "signature", paramKey: "signature"},
			}

			// 添加表单字段
			for _, field := range formFields {
				var fieldValue string
				switch field.key {
				case "key":
					fieldValue = objectKey
				case "success_action_status":
					fieldValue = "200"
				default:
					if field.paramKey != "" {
						if val, ok := uploadParams[field.paramKey]; ok {
							fieldValue = fmt.Sprintf("%v", val)
						} else {
							log.Printf("警告: 缺少字段 %s (参数名: %s)", field.key, field.paramKey)
							continue
						}
					}
				}
				if err := w.WriteField(field.key, fieldValue); err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("写入通知ID %d 的表单字段 %s 失败: %v", oldNotice.ID, field.key, err))
					continue
				}
				log.Printf("添加字段: %s = %s", field.key, fieldValue)
			}

			// 最后添加文件数据
			fw, err := w.CreateFormFile("file", objectKey)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("创建通知ID %d 的文件表单失败: %v", oldNotice.ID, err))
				continue
			}
			if _, err = io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("复制通知ID %d 的文件数据失败: %v", oldNotice.ID, err))
				continue
			}
			w.Close()

			// 上传到OSS
			uploadReq, err := http.NewRequest("POST", uploadParams["host"].(string), &b)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("创建通知ID %d 的上传请求失败: %v", oldNotice.ID, err))
				continue
			}
			uploadReq.Header.Set("Content-Type", w.FormDataContentType())

			// 添加调试日志
			log.Printf("上传URL: %s", uploadParams["host"].(string))
			log.Printf("上传参数: %+v", uploadParams)
			log.Printf("上传文件名: %s", objectKey)
			log.Printf("Content-Type: %s", w.FormDataContentType())

			client := &http.Client{}
			uploadResp, err := client.Do(uploadReq)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("上传通知ID %d 的文件失败: %v", oldNotice.ID, err))
				continue
			}

			// 读取响应
			respBody, _ := io.ReadAll(uploadResp.Body)
			uploadResp.Body.Close()

			if uploadResp.StatusCode != http.StatusOK {
				failedNotices = append(failedNotices, fmt.Sprintf("通知ID %d 的文件上传失败，状态码: %d, 响应: %s", oldNotice.ID, uploadResp.StatusCode, string(respBody)))
				// 删除已创建的文件记录
				if shouldAddFile {
					if err := databases.DB_CONN.Delete(fileForNotice).Error; err != nil {
						log.Printf("删除失败的文件记录失败: %v", err)
					}
				}
				continue
			}

			// 验证上传是否成功
			log.Printf("文件上传响应: %s", string(respBody))
		}

		// 确定通知类型
		noticeType := field.NoticeTypeNormal
		switch oldNotice.MessType {
		case "urgent":
			noticeType = field.NoticeTypeUrgent
		case "normal":
			noticeType = field.NoticeTypeNormal
		case "building":
			noticeType = field.NoticeTypeBuilding
		case "government":
			noticeType = field.NoticeTypeGovernment
		default:
			noticeType = field.NoticeTypeNormal
		}

		// 创建通知记录
		notice := &base_models.Notice{
			Title:       oldNotice.MessTitle,
			Description: oldNotice.MessTitle,
			Type:        noticeType,
			Status:      field.Status("active"),
			StartTime:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			EndTime:     time.Date(2100, 2, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			IsPublic:    true,
			FileID:      &fileForNotice.ID,
			FileType:    field.FileTypePdf,
		}

		// 创建通知并绑定到建筑物
		err = databases.DB_CONN.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(notice).Error; err != nil {
				return fmt.Errorf("创建通知记录失败: %v", err)
			}

			if err := tx.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)", notice.ID, id).Error; err != nil {
				return fmt.Errorf("绑定通知到建筑物失败: %v", err)
			}

			return nil
		})
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("处理通知ID %d 失败: %v", oldNotice.ID, err))
			continue
		}

		successCount++
	}

	// 返回同步结果
	c.Ctx.JSON(200, gin.H{
		"message":        "同步完成",
		"successCount":   successCount,
		"hasSyncedCount": hasSyncedCount,
		"failedNotices":  failedNotices,
		"totalProcessed": len(respBody),
	})
}
