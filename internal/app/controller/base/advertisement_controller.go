package http_base_controller

import (
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	databases "github.com/The-Healthist/iboard_http_service/internal/infrastructure/database"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/The-Healthist/iboard_http_service/pkg/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

// AdvertisementController handles advertisement operations
type AdvertisementController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

// NewAdvertisementController creates a new advertisement controller
func NewAdvertisementController(ctx *gin.Context, container *container.ServiceContainer) *AdvertisementController {
	return &AdvertisementController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncAdvertisement returns a gin.HandlerFunc for the specified method
func HandleFuncAdvertisement(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewAdvertisementController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

type CreateAdvertisementRequest struct {
	Title       string                     `json:"title" binding:"required" example:"夏季促销活动"`
	Description string                     `json:"description" example:"夏季促销活动，全场8折"`
	Type        field.AdvertisementType    `json:"type" binding:"required" example:"image"`
	Status      field.Status               `json:"status" binding:"required" example:"active"`
	Duration    int                        `json:"duration" example:"30"`
	Priority    int                        `json:"priority" example:"1"`
	StartTime   *time.Time                 `json:"startTime" binding:"required" example:"2023-06-01T00:00:00Z"`
	EndTime     *time.Time                 `json:"endTime" binding:"required" example:"2023-08-31T23:59:59Z"`
	Display     field.AdvertisementDisplay `json:"display" binding:"required" example:"fullscreen"`
	IsPublic    bool                       `json:"isPublic" example:"true"`
	Path        string                     `json:"path" binding:"required" example:"/uploads/images/summer_sale.jpg"`
}

// 1.Create 创建广告
// @Summary      创建广告
// @Description  创建一个新的广告
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        advertisement body CreateAdvertisementRequest true "广告信息"
// @Success      200  {object}  map[string]interface{} "返回创建的广告信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/advertisement [post]
// @Security     BearerAuth
func (c *AdvertisementController) Create() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	var form CreateAdvertisementRequest
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		log.Warn("创建广告表单无效 | %v | 错误: %v", requestID, err)
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	log.Info("尝试创建广告 | %v | 标题: %s", requestID, form.Title)

	// Check if file exists
	var file base_models.File
	if form.Path != "" {
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			log.Warn("创建广告失败，文件不存在 | %v | 路径: %s", requestID, form.Path)
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
	}

	advertisement := &base_models.Advertisement{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		Status:      form.Status,
		Duration:    form.Duration,
		Priority:    form.Priority,
		StartTime:   *form.StartTime, // 使用指针值
		EndTime:     *form.EndTime,   // 使用指针值
		Display:     form.Display,
		IsPublic:    form.IsPublic,
	}

	// Set file ID if file exists
	if file.ID != 0 {
		advertisement.FileID = &file.ID
	}

	if err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Create(advertisement); err != nil {
		log.Error("创建广告失败 | %v | 标题: %s | 错误: %v", requestID, form.Title, err)
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create advertisement failed",
		})
		return
	}

	log.Info("创建广告成功 | %v | 广告ID: %d", requestID, advertisement.ID)
	c.Ctx.JSON(200, gin.H{
		"message": "create advertisement success",
		"data":    advertisement,
	})
}

// 2.CreateMany 批量创建广告
// @Summary      批量创建广告
// @Description  批量创建多个广告
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        advertisements body []CreateAdvertisementRequest true "广告信息数组"
// @Success      200  {object}  map[string]interface{} "返回创建的广告信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/advertisements [post]
// @Security     BearerAuth
func (c *AdvertisementController) CreateMany() {
	var forms []struct {
		Title       string                     `json:"title" binding:"required" example:"夏季促销活动"`
		Description string                     `json:"description" example:"夏季促销活动，全场8折"`
		Type        field.AdvertisementType    `json:"type" binding:"required" example:"image"`
		Status      field.Status               `json:"status" binding:"required" example:"active"`
		Duration    int                        `json:"duration" example:"30"`
		Priority    int                        `json:"priority" binding:"required" example:"1"`
		StartTime   *time.Time                 `json:"startTime" binding:"required" example:"2023-06-01T00:00:00Z"`
		EndTime     *time.Time                 `json:"endTime" binding:"required" example:"2023-08-31T23:59:59Z"`
		Display     field.AdvertisementDisplay `json:"display" binding:"required" example:"fullscreen"`
		IsPublic    bool                       `json:"isPublic" example:"true"`
		Path        string                     `json:"path" binding:"required" example:"/uploads/images/summer_sale.jpg"`
	}

	if err := c.Ctx.ShouldBindJSON(&forms); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var advertisements []*base_models.Advertisement
	for _, form := range forms {
		// If path is provided, find the corresponding file
		var fileID *uint
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
		}

		advertisement := &base_models.Advertisement{
			Title:       form.Title,
			Description: form.Description,
			Type:        form.Type,
			Status:      form.Status,
			Duration:    form.Duration,
			Priority:    form.Priority,
			StartTime:   *form.StartTime,
			EndTime:     *form.EndTime,
			Display:     form.Display,
			IsPublic:    form.IsPublic,
			FileID:      fileID,
		}
		advertisements = append(advertisements, advertisement)
	}

	for _, advertisement := range advertisements {
		if err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Create(advertisement); err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":         err.Error(),
				"message":       "create advertisement failed",
				"advertisement": advertisement,
			})
			return
		}
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create advertisements success",
		"data":    advertisements,
	})
}

// 3.Get 获取广告列表
// @Summary      获取广告列表
// @Description  根据查询条件获取广告列表
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        search query string false "搜索关键词" example:"促销"
// @Param        type query string false "广告类型" example:"image"
// @Param        pageSize query int false "每页数量" default(10)
// @Param        pageNum query int false "页码" default(1)
// @Param        desc query bool false "是否降序" default(false)
// @Success      200  {object}  map[string]interface{} "返回广告列表和分页信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/advertisement [get]
// @Security     BearerAuth
func (c *AdvertisementController) Get() {
	var searchQuery struct {
		Search string `form:"search"`
		Type   string `form:"type"`
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

	advertisements, paginationResult, err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       advertisements,
		"pagination": paginationResult,
	})
}

// 4.Update 更新广告
// @Summary      更新广告
// @Description  更新广告信息
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        advertisement body object true "广告更新信息"
// @Param        id formData uint true "广告ID" example:"1"
// @Param        title formData string false "广告标题" example:"秋季促销活动"
// @Param        description formData string false "广告描述" example:"秋季促销活动，全场7折"
// @Param        type formData string false "广告类型" example:"video"
// @Param        status formData string false "状态" example:"active"
// @Param        duration formData int false "持续时间(秒)" example:"45"
// @Param        priority formData int false "优先级" example:"2"
// @Param        startTime formData string false "开始时间" example:"2023-09-01T00:00:00Z"
// @Param        endTime formData string false "结束时间" example:"2023-11-30T23:59:59Z"
// @Param        display formData string false "显示方式" example:"popup"
// @Param        isPublic formData bool false "是否公开" example:"false"
// @Param        path formData string false "文件路径" example:"/uploads/videos/autumn_sale.mp4"
// @Success      200  {object}  map[string]interface{} "返回更新后的广告信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/advertisement [put]
// @Security     BearerAuth
func (c *AdvertisementController) Update() {
	var form struct {
		ID          uint                       `json:"id" binding:"required" example:"1"`
		Title       string                     `json:"title" example:"秋季促销活动"`
		Description string                     `json:"description" example:"秋季促销活动，全场7折"`
		Type        field.AdvertisementType    `json:"type" example:"video"`
		Status      field.Status               `json:"status" example:"active"`
		Duration    *int                       `json:"duration" example:"45"`
		Priority    *int                       `json:"priority" example:"2"`
		StartTime   *time.Time                 `json:"startTime" example:"2023-09-01T00:00:00Z"`
		EndTime     *time.Time                 `json:"endTime" example:"2023-11-30T23:59:59Z"`
		Display     field.AdvertisementDisplay `json:"display" example:"popup"`
		IsPublic    *bool                      `json:"isPublic" example:"false"`
		Path        string                     `json:"path" example:"/uploads/videos/autumn_sale.mp4"`
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
	if form.Duration != nil {
		updates["duration"] = *form.Duration
	}
	if form.Priority != nil {
		updates["priority"] = *form.Priority
	}
	if form.StartTime != nil {
		updates["start_time"] = form.StartTime
	}
	if form.EndTime != nil {
		updates["end_time"] = form.EndTime
	}
	if form.Display != "" {
		updates["display"] = form.Display
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
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

	advertisement, err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update advertisement success",
		"data":    advertisement,
	})
}

// 5.Delete 删除广告
// @Summary      删除广告
// @Description  删除一个或多个广告
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        ids body object true "广告ID列表"
// @Param        ids.ids body []uint true "广告ID数组" example:"[1,2,3]"
// @Success      200  {object}  map[string]interface{} "删除成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/advertisement [delete]
// @Security     BearerAuth
func (c *AdvertisementController) Delete() {
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

	// 1. 获取所有要删除的广告信息
	var advertisements []base_models.Advertisement
	if err := tx.Where("id IN ?", form.IDs).Find(&advertisements).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get advertisements",
			"message": err.Error(),
		})
		return
	}

	// 2. 解除与建筑物的关联
	if err := tx.Exec("DELETE FROM advertisement_buildings WHERE advertisement_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind buildings",
			"message": err.Error(),
		})
		return
	}

	// 3. 收集所有关联的文件ID
	var fileIDs []uint
	for _, ad := range advertisements {
		if ad.FileID != nil {
			fileIDs = append(fileIDs, *ad.FileID)
		}
	}

	// 4. 解除文件绑定
	if err := tx.Model(&base_models.Advertisement{}).Where("id IN ?", form.IDs).Update("file_id", nil).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind files",
			"message": err.Error(),
		})
		return
	}

	// 5. 检查每个文件的引用并删除未被引用的文件
	for _, fileID := range fileIDs {
		var adCount int64
		var noticeCount int64

		if err := tx.Model(&base_models.Advertisement{}).Where("file_id = ?", fileID).Count(&adCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check advertisement references",
				"message": err.Error(),
			})
			return
		}

		if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", fileID).Count(&noticeCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check notice references",
				"message": err.Error(),
			})
			return
		}

		// 如果没有其他引用，删除文件
		if adCount == 0 && noticeCount == 0 {
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

	// 6. 删除广告
	if err := tx.Delete(&base_models.Advertisement{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete advertisements",
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

	c.Ctx.JSON(200, gin.H{"message": "delete advertisement success"})
}

// 6.GetOne 获取单个广告
// @Summary      获取单个广告
// @Description  根据ID获取广告详细信息
// @Tags         Advertisement
// @Accept       json
// @Produce      json
// @Param        id path int true "广告ID" example:"1"
// @Success      200  {object}  map[string]interface{} "返回广告详细信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "广告不存在"
// @Router       /admin/advertisement/{id} [get]
// @Security     BearerAuth
func (c *AdvertisementController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	advertisement, err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get advertisement success",
		"data":    advertisement,
	})
}
