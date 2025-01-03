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

	// start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. get all buildings to be deleted
	var buildings []base_models.Building
	if err := tx.Preload("Notices").Preload("Advertisements").Where("id IN ?", form.IDs).Find(&buildings).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get buildings",
			"message": err.Error(),
		})
		return
	}

	// 2. collect all file ids of notices and advertisements
	fileIDMap := make(map[uint]bool)
	var noticeIDs []uint
	var advertisementIDs []uint
	for _, building := range buildings {
		for _, notice := range building.Notices {
			if notice.FileID != nil {
				fileIDMap[*notice.FileID] = true
			}
			noticeIDs = append(noticeIDs, notice.ID)
		}
		for _, ad := range building.Advertisements {
			if ad.FileID != nil {
				fileIDMap[*ad.FileID] = true
			}
			advertisementIDs = append(advertisementIDs, ad.ID)
		}
	}

	// convert map to slice
	var fileIDs []uint
	for fileID := range fileIDMap {
		fileIDs = append(fileIDs, fileID)
	}

	// 3. unbind notices
	if len(noticeIDs) > 0 {
		// unbind notice-building
		if err := tx.Exec("DELETE FROM notice_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to unbind notices from buildings",
				"message": err.Error(),
			})
			return
		}
	}

	// 4. unbind advertisements
	if len(advertisementIDs) > 0 {
		// unbind advertisement-building
		if err := tx.Exec("DELETE FROM advertisement_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to unbind advertisements from buildings",
				"message": err.Error(),
			})
			return
		}
	}

	// 5. unbind admins
	if err := tx.Exec("DELETE FROM building_admins_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind admins",
			"message": err.Error(),
		})
		return
	}

	// 6. delete buildings
	if err := tx.Delete(&base_models.Building{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete buildings",
			"message": err.Error(),
		})
		return
	}

	// commit transaction
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
	// get building id from url param
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building ID",
			"message": "Please check the ID format",
		})
		return
	}

	// get building info to get ismartId
	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get building information",
			"message": err.Error(),
		})
		return
	}

	// prepare request to old system api
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to prepare request data",
			"message": err.Error(),
		})
		return
	}

	// request old system to get notice list
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get notices from old system",
			"message": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// parse response data
	var respBody []struct {
		ID        int    `json:"id"`
		MessTitle string `json:"mess_title"`
		MessType  string `json:"mess_type"`
		MessFile  string `json:"mess_file"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to parse response data",
			"message": err.Error(),
		})
		return
	}

	// process each notice
	var successCount int
	var failedNotices []string
	var hasSyncedCount int

	for _, oldNotice := range respBody {
		// download pdf file
		fileResp, err := http.Get(oldNotice.MessFile)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to download file for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		// read file content
		fileContent, err := io.ReadAll(fileResp.Body)
		fileResp.Body.Close()
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to read file content for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		fileSize := len(fileContent)
		md5Hash := md5.Sum(fileContent)
		md5Str := hex.EncodeToString(md5Hash[:])
		mimeType := http.DetectContentType(fileContent)

		// get uploader info from token
		claims, exists := c.Ctx.Get("claims")
		if !exists {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: token claims not found", oldNotice.ID))
			continue
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: invalid claims format", oldNotice.ID))
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
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: invalid uploader info", oldNotice.ID))
			continue
		}

		// generate file name and get upload params
		currentTime := time.Now()
		dir := currentTime.Format("2006-01-02") + "/"
		fileName := fmt.Sprintf("%s.pdf", uuid.New().String())
		objectKey := dir + fileName
		uploadParams, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParamsSync(objectKey)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get upload params for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		// check if md5 exists before create file record
		var existingFile base_models.File
		var shouldUpload bool = true
		var shouldAddFile bool = true
		var fileForNotice *base_models.File

		err = databases.DB_CONN.Where("md5 = ?", md5Str).First(&existingFile).Error
		if err == nil {
			// file exists, check if it is associated with notice
			var notice base_models.Notice
			var noticeExists bool

			// check if file is associated with notice
			err := databases.DB_CONN.Where("file_id = ?", existingFile.ID).First(&notice).Error
			noticeExists = err == nil

			if noticeExists {
				// check if notice is associated with current building
				var count int64
				err = databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ? AND building_id = ?", notice.ID, id).
					Count(&count).Error
				if err == nil && count > 0 {
					// notice is associated with current building, skip processing
					hasSyncedCount++
					log.Printf("Notice ID %d has already been synced, skipping", oldNotice.ID)
					continue
				}

				if err := databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ?", notice.ID).
					Count(&count).Error; err == nil && count > 0 {
					// notice is associated with other building, need to create new file and notice
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
					// notice exists but not associated with any building, bind to current building only
					shouldUpload = false
					shouldAddFile = false
					fileForNotice = &existingFile

					// bind notice to current building directly
					if err := databases.DB_CONN.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)", notice.ID, id).Error; err != nil {
						failedNotices = append(failedNotices, fmt.Sprintf("Failed to bind notice ID %d to building: %v", oldNotice.ID, err))
						continue
					}
					successCount++
					continue
				}
			} else {
				// file exists but not associated with notice, use existing file to create new notice
				shouldUpload = false
				shouldAddFile = false
				fileForNotice = &existingFile
			}
		} else if err == gorm.ErrRecordNotFound {
			// file not exists, need to create all
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
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to check MD5 for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		// If need to create new file
		if shouldAddFile {
			if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(fileForNotice); err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to create file record for notice ID %d: %v", oldNotice.ID, err))
				continue
			}
		}

		// If need to upload file
		if shouldUpload {
			// Prepare upload form data
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			// Add form fields in the exact same order as frontend
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

			// Add form fields
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
							log.Printf("Warning: missing field %s (param name: %s)", field.key, field.paramKey)
							continue
						}
					}
				}
				if err := w.WriteField(field.key, fieldValue); err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to write form field %s for notice ID %d: %v", field.key, oldNotice.ID, err))
					continue
				}
				log.Printf("Added field: %s = %s", field.key, fieldValue)
			}

			// Finally add file data
			fw, err := w.CreateFormFile("file", objectKey)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to create file form for notice ID %d: %v", oldNotice.ID, err))
				continue
			}
			if _, err = io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to copy file data for notice ID %d: %v", oldNotice.ID, err))
				continue
			}
			w.Close()

			// Upload to OSS
			uploadReq, err := http.NewRequest("POST", uploadParams["host"].(string), &b)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to create upload request for notice ID %d: %v", oldNotice.ID, err))
				continue
			}
			uploadReq.Header.Set("Content-Type", w.FormDataContentType())

			// Add debug logs
			log.Printf("Upload URL: %s", uploadParams["host"].(string))
			log.Printf("Upload params: %+v", uploadParams)
			log.Printf("Upload file name: %s", objectKey)
			log.Printf("Content-Type: %s", w.FormDataContentType())

			client := &http.Client{}

			uploadResp, err := client.Do(uploadReq)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to upload file for notice ID %d: %v", oldNotice.ID, err))
				continue
			}

			// Read response
			respBody, _ := io.ReadAll(uploadResp.Body)
			uploadResp.Body.Close()

			if uploadResp.StatusCode != http.StatusOK {
				failedNotices = append(failedNotices, fmt.Sprintf("File upload failed for notice ID %d: status code %d, response: %s", oldNotice.ID, uploadResp.StatusCode, string(respBody)))
				// Delete created file record if upload failed
				if shouldAddFile {
					if err := databases.DB_CONN.Delete(fileForNotice).Error; err != nil {
						log.Printf("Failed to delete failed file record: %v", err)
					}
				}
				continue
			}

			// Verify upload success
			log.Printf("File upload response: %s", string(respBody))
		}

		// Determine notice type
		noticeType := field.NoticeTypeNormal
		switch oldNotice.MessType {
		case string(field.NoticeOldTypeCommon):
			noticeType = field.NoticeTypeNormal
		case string(field.NoticeOldTypeIo):
			noticeType = field.NoticeTypeBuilding
		case string(field.NoticeOldTypeUrgent):
			noticeType = field.NoticeTypeUrgent
		case string(field.NoticeOldTypeGovernment):
			noticeType = field.NoticeTypeGovernment
		default:
			noticeType = field.NoticeTypeNormal
		}

		// Create notice record
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

		// Create notice and bind to building
		err = databases.DB_CONN.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(notice).Error; err != nil {
				return fmt.Errorf("failed to create notice record: %v", err)
			}

			if err := tx.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)", notice.ID, id).Error; err != nil {
				return fmt.Errorf("failed to bind notice to building: %v", err)
			}

			return nil
		})
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		successCount++
	}

	// Return sync results
	c.Ctx.JSON(200, gin.H{
		"message":        "Sync completed",
		"successCount":   successCount,
		"hasSyncedCount": hasSyncedCount,
		"failedNotices":  failedNotices,
		"totalProcessed": len(respBody),
	})
}
