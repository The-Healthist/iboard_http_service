package http_base_controller

import (
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	databases "github.com/The-Healthist/iboard_http_service/internal/infrastructure/database"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceNoticeController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

// NoticeController handles notice operations
type NoticeController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewNoticeController creates a new notice controller
func NewNoticeController(ctx *gin.Context, container *container.ServiceContainer) *NoticeController {
	return &NoticeController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncNotice returns a gin.HandlerFunc for the specified method
func HandleFuncNotice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

type CreateNoticeRequest struct {
	Title          string           `json:"title" binding:"required" example:"业主大会通知"`
	Description    string           `json:"description" example:"关于召开2023年度业主大会的通知"`
	Type           field.NoticeType `json:"type" binding:"required" example:"normal"`
	Status         field.Status     `json:"status" binding:"required" example:"active"`
	StartTime      *time.Time       `json:"startTime" binding:"required" example:"2023-06-01T00:00:00Z"`
	EndTime        *time.Time       `json:"endTime" binding:"required" example:"2023-06-30T23:59:59Z"`
	IsPublic       bool             `json:"isPublic" example:"true"`
	IsIsmartNotice bool             `json:"isIsmartNotice" example:"false"`
	Priority       int              `json:"priority" example:"1"`
	Path           string           `json:"path" binding:"required" example:"/uploads/documents/owners_meeting.pdf"`
	FileType       field.FileType   `json:"fileType" example:"pdf"`
}

// 1.Create 创建通知
// @Summary      创建通知
// @Description  创建一个新的通知
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        notice body CreateNoticeRequest true "通知信息"
// @Success      200  {object}  map[string]interface{} "返回创建的通知信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/notice [post]
// @Security     BearerAuth
func (c *NoticeController) Create() {
	var form CreateNoticeRequest

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Set default values for startTime, endTime, and status
	now := time.Now()
	startTime := now
	endTime := now.AddDate(1, 0, 0) // 1 year from now
	status := field.Status("active")

	if form.StartTime != nil {
		startTime = *form.StartTime
	}
	if form.EndTime != nil {
		endTime = *form.EndTime
	}
	if form.Status != "" {
		status = form.Status
	}

	// If path is provided, find the corresponding file
	var fileID *uint
	fileType := form.FileType
	if form.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID

		// If FileType is not provided, infer it from the file path
		if fileType == "" {
			// Extract file extension from path
			path := form.Path
			if len(path) > 0 {
				// Find the last dot in the path
				lastDot := -1
				for i := len(path) - 1; i >= 0; i-- {
					if path[i] == '.' {
						lastDot = i
						break
					}
				}
				if lastDot != -1 && lastDot < len(path)-1 {
					ext := path[lastDot+1:]
					// Convert to lowercase and check if it's a valid file type
					switch ext {
					case "pdf":
						fileType = field.FileTypePdf
					default:
						// Default to pdf if extension is not recognized
						fileType = field.FileTypePdf
					}
				} else {
					// Default to pdf if no extension found
					fileType = field.FileTypePdf
				}
			} else {
				// Default to pdf if path is empty
				fileType = field.FileTypePdf
			}
		}
	} else {
		// If no path provided, default to pdf
		if fileType == "" {
			fileType = field.FileTypePdf
		}
	}

	notice := &base_models.Notice{
		Title:          form.Title,
		Description:    form.Description,
		Type:           form.Type,
		Status:         status,
		StartTime:      startTime,
		EndTime:        endTime,
		FileID:         fileID,
		IsPublic:       form.IsPublic,
		IsIsmartNotice: form.IsIsmartNotice,
		Priority:       form.Priority,
		FileType:       fileType,
	}

	if err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Create(notice); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create notice failed",
		})
		return
	}

	// Reload notice to get associated file information
	if err := databases.DB_CONN.Preload("File").First(notice, notice.ID).Error; err != nil {
		c.Ctx.JSON(200, gin.H{
			"message": "create notice success, but failed to load file info",
			"data":    notice,
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create notice success",
		"data":    notice,
	})
}

// 2.CreateMany 批量创建通知
// @Summary      批量创建通知
// @Description  批量创建多个通知
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        notices body []CreateNoticeRequest true "通知信息数组"
// @Success      200  {object}  map[string]interface{} "返回创建的通知信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/notices [post]
// @Security     BearerAuth
func (c *NoticeController) CreateMany() {
	var forms []struct {
		Title          string           `json:"title" binding:"required" example:"业主大会通知"`
		Description    string           `json:"description" example:"关于召开2023年度业主大会的通知"`
		Type           field.NoticeType `json:"type" binding:"required" example:"normal"`
		Status         field.Status     `json:"status" binding:"required" example:"active"`
		StartTime      *time.Time       `json:"startTime" binding:"required" example:"2023-06-01T00:00:00Z"`
		EndTime        *time.Time       `json:"endTime" binding:"required" example:"2023-06-30T23:59:59Z"`
		IsPublic       bool             `json:"isPublic" example:"true"`
		IsIsmartNotice bool             `json:"isIsmartNotice" example:"false"`
		Priority       int              `json:"priority" example:"1"`
		Path           string           `json:"path" binding:"required" example:"/uploads/documents/owners_meeting.pdf"`
		FileType       field.FileType   `json:"fileType" example:"pdf"`
	}

	if err := c.Ctx.ShouldBindJSON(&forms); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var notices []*base_models.Notice
	for _, form := range forms {
		// Handle FileType similar to single create
		fileType := form.FileType
		if fileType == "" {
			// Default to pdf if not provided
			fileType = field.FileTypePdf
		}

		notice := &base_models.Notice{
			Title:          form.Title,
			Description:    form.Description,
			Type:           form.Type,
			Status:         form.Status,
			StartTime:      *form.StartTime,
			EndTime:        *form.EndTime,
			IsPublic:       form.IsPublic,
			IsIsmartNotice: form.IsIsmartNotice,
			Priority:       form.Priority,
			FileType:       fileType,
		}
		notices = append(notices, notice)
	}

	for _, notice := range notices {
		if err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Create(notice); err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "create notice failed",
				"notice":  notice,
			})
			return
		}
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create notices success",
		"data":    notices,
	})
}

// 3.Get 获取通知列表
// @Summary      获取通知列表
// @Description  根据查询条件获取通知列表
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        search query string false "搜索关键词" example:"业主大会"
// @Param        type query string false "通知类型" example:"normal"
// @Param        pageSize query int false "每页数量" default(10)
// @Param        pageNum query int false "页码" default(1)
// @Param        desc query bool false "是否降序" default(false)
// @Success      200  {object}  map[string]interface{} "返回通知列表和分页信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/notice [get]
// @Security     BearerAuth
func (c *NoticeController) Get() {
	var searchQuery struct {
		Search string `form:"search" example:"业主大会"`
		Type   string `form:"type" example:"normal"`
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

	notices, paginationResult, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": paginationResult,
	})
}

// 4.Update 更新通知
// @Summary      更新通知
// @Description  更新通知信息
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        notice body object true "通知更新信息"
// @Param        id formData uint true "通知ID" example:"1"
// @Param        title formData string false "通知标题" example:"业主大会通知（更新）"
// @Param        description formData string false "通知描述" example:"关于召开2023年度业主大会的通知（日期更新）"
// @Param        type formData string false "通知类型" example:"urgent"
// @Param        status formData string false "状态" example:"active"
// @Param        startTime formData string false "开始时间" example:"2023-07-01T00:00:00Z"
// @Param        endTime formData string false "结束时间" example:"2023-07-31T23:59:59Z"
// @Param        isPublic formData bool false "是否公开" example:"true"
// @Param        priority formData int false "优先级" example:"2"
// @Param        path formData string false "文件路径" example:"/uploads/documents/owners_meeting_updated.pdf"
// @Param        fileType formData string false "文件类型" example:"pdf"
// @Success      200  {object}  map[string]interface{} "返回更新后的通知信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/notice [put]
// @Security     BearerAuth
func (c *NoticeController) Update() {
	var form struct {
		ID          uint             `json:"id" binding:"required" example:"1"`
		Title       string           `json:"title" example:"业主大会通知（更新）"`
		Description string           `json:"description" example:"关于召开2023年度业主大会的通知（日期更新）"`
		Type        field.NoticeType `json:"type" example:"urgent"`
		Status      field.Status     `json:"status" example:"active"`
		StartTime   *time.Time       `json:"startTime" example:"2023-07-01T00:00:00Z"`
		EndTime     *time.Time       `json:"endTime" example:"2023-07-31T23:59:59Z"`
		IsPublic    *bool            `json:"isPublic" example:"true"`
		Priority    *int             `json:"priority" example:"2"`
		Path        string           `json:"path" example:"/uploads/documents/owners_meeting_updated.pdf"`
		FileType    field.FileType   `json:"fileType" example:"pdf"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.Type != "" {
		updates["type"] = form.Type
	}
	if form.Status != "" {
		updates["status"] = form.Status
	}
	if form.StartTime != nil {
		updates["start_time"] = form.StartTime
	}
	if form.EndTime != nil {
		updates["end_time"] = form.EndTime
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}
	if form.Priority != nil {
		updates["priority"] = *form.Priority
	}
	if form.FileType != "" {
		updates["file_type"] = form.FileType
	}

	// If new path is provided
	if form.Path != "" {
		// Start transaction
		tx := databases.DB_CONN.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Find new file
		var newFile base_models.File
		if err := tx.Where("path = ?", form.Path).First(&newFile).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "File not found",
				"message": err.Error(),
			})
			return
		}

		updates["file_id"] = newFile.ID
		if err := tx.Commit().Error; err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to update file",
				"message": err.Error(),
			})
			return
		}
	}

	notice, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update notice success",
		"data":    notice,
	})
}

// 5.Delete 删除通知
// @Summary      删除通知
// @Description  删除一个或多个通知
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        ids body object true "通知ID列表"
// @Param        ids.ids body []uint true "通知ID数组" example:"[1,2,3]"
// @Success      200  {object}  map[string]interface{} "删除成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/notice [delete]
// @Security     BearerAuth
func (c *NoticeController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required" example:"[1,2,3]"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 开启事务
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取所有要删除的通知信息
	var notices []base_models.Notice
	if err := tx.Where("id IN ?", form.IDs).Find(&notices).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get notices",
			"message": err.Error(),
		})
		return
	}

	// 2. 解除与建筑物的关联
	if err := tx.Exec("DELETE FROM notice_buildings WHERE notice_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind buildings",
			"message": err.Error(),
		})
		return
	}

	// 3. 收集所有关联的文件ID
	var fileIDs []uint
	for _, notice := range notices {
		if notice.FileID != nil {
			fileIDs = append(fileIDs, *notice.FileID)
		}
	}

	// 4. 解除文件绑定
	if err := tx.Model(&base_models.Notice{}).Where("id IN ?", form.IDs).Update("file_id", nil).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind files",
			"message": err.Error(),
		})
		return
	}

	// 5. 检查每个文件的引用并删除未被引用的文件
	for _, fileID := range fileIDs {
		var noticeCount int64
		var adCount int64

		if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", fileID).Count(&noticeCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check notice references",
				"message": err.Error(),
			})
			return
		}

		if err := tx.Model(&base_models.Advertisement{}).Where("file_id = ?", fileID).Count(&adCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check advertisement references",
				"message": err.Error(),
			})
			return
		}

		// 如果没有其他引用，删除文件
		if noticeCount == 0 && adCount == 0 {
			if err := tx.Delete(&base_models.File{}, fileID).Error; err != nil {
				tx.Rollback()
				c.Ctx.JSON(400, gin.H{
					"error":   "Failed to delete file",
					"message": err.Error(),
				})
				return
			}
		}
	}

	// 6. 删除通知
	if err := tx.Delete(&base_models.Notice{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete notices",
			"message": err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "failed to commit transaction",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete notice success"})
}

// 6.GetOne 获取单个通知
// @Summary      获取单个通知
// @Description  根据ID获取通知详细信息
// @Tags         Notice
// @Accept       json
// @Produce      json
// @Param        id path int true "通知ID" example:"1"
// @Success      200  {object}  map[string]interface{} "返回通知详细信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "通知不存在"
// @Router       /admin/notice/{id} [get]
// @Security     BearerAuth
func (c *NoticeController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid notice ID"})
		return
	}

	notice, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get notice success",
		"data":    notice,
	})
}
