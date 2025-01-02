package http_base_controller

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
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

// Create creates a new advertisement
func (c *AdvertisementController) Create() {
	var form struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		Type        field.AdvertisementType    `json:"type"`
		Status      field.Status               `json:"status"`
		Duration    int                        `json:"duration"`
		StartTime   *time.Time                 `json:"startTime"`
		EndTime     *time.Time                 `json:"endTime"`
		Display     field.AdvertisementDisplay `json:"display"`
		IsPublic    bool                       `json:"isPublic"`
		Path        string                     `json:"path"`
	}

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

	// Start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// If path is provided, find the corresponding file
	var fileID *uint
	if form.Path != "" {
		var file base_models.File
		if err := tx.Where("path = ?", form.Path).First(&file).Error; err != nil {
			tx.Rollback()
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
		Status:      status,
		Duration:    form.Duration,
		StartTime:   startTime,
		EndTime:     endTime,
		Display:     form.Display,
		FileID:      fileID,
		IsPublic:    form.IsPublic,
	}

	if err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Create(advertisement); err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create advertisement failed",
		})
		return
	}

	// Reload advertisement to get associated file information
	if err := tx.Preload("File").First(advertisement, advertisement.ID).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(200, gin.H{
			"message": "create advertisement success, but failed to load file info",
			"data":    advertisement,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "failed to commit transaction",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create advertisement success",
		"data":    advertisement,
	})
}

// CreateMany creates multiple advertisements
func (c *AdvertisementController) CreateMany() {
	var forms []struct {
		Title       string                     `json:"title" binding:"required"`
		Description string                     `json:"description"`
		Type        field.AdvertisementType    `json:"type"`
		Status      field.Status               `json:"status"`
		Duration    int                        `json:"duration"`
		StartTime   *time.Time                 `json:"startTime"`
		EndTime     *time.Time                 `json:"endTime"`
		Display     field.AdvertisementDisplay `json:"display"`
		IsPublic    bool                       `json:"isPublic"`
		FileID      *uint                      `json:"fileId"`
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
		advertisement := &base_models.Advertisement{
			Title:       form.Title,
			Description: form.Description,
			Type:        form.Type,
			Status:      form.Status,
			Duration:    form.Duration,
			StartTime:   *form.StartTime,
			EndTime:     *form.EndTime,
			Display:     form.Display,
			FileID:      form.FileID,
			IsPublic:    form.IsPublic,
		}
		advertisements = append(advertisements, advertisement)
	}

	// Start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, advertisement := range advertisements {
		if err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Create(advertisement); err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":         err.Error(),
				"message":       "create advertisement failed",
				"advertisement": advertisement,
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "failed to commit transaction",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create advertisements success",
		"data":    advertisements,
	})
}

// Get retrieves advertisements based on search criteria
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

// Update updates an advertisement
func (c *AdvertisementController) Update() {
	var form struct {
		ID          uint                       `json:"id" binding:"required"`
		Title       string                     `json:"title"`
		Description string                     `json:"description"`
		Type        field.AdvertisementType    `json:"type"`
		Status      field.Status               `json:"status"`
		StartTime   *time.Time                 `json:"startTime"`
		EndTime     *time.Time                 `json:"endTime"`
		Display     field.AdvertisementDisplay `json:"display"`
		IsPublic    *bool                      `json:"isPublic"`
		Path        string                     `json:"path"`
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
	if form.Display != "" {
		updates["display"] = form.Display
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}

	// If new path is provided
	if form.Path != "" {
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
	}

	advertisement, err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Update(form.ID, updates)
	if err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "update advertisement failed",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "failed to commit transaction",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update advertisement success",
		"data":    advertisement,
	})
}

// Delete deletes advertisements
func (c *AdvertisementController) Delete() {
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

	// Get all advertisements to be deleted
	var advertisements []base_models.Advertisement
	if err := tx.Where("id IN ?", form.IDs).Find(&advertisements).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get advertisements",
			"message": err.Error(),
		})
		return
	}
	// TODO:
	// Collect all file IDs
	// var fileIDs []uint
	// for _, ad := range advertisements {
	// 	if ad.FileID != nil {
	// 		fileIDs = append(fileIDs, *ad.FileID)
	// 	}
	// }

	// Delete advertisement-building associations
	if err := tx.Exec("DELETE FROM advertisement_buildings WHERE advertisement_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete advertisement-building associations",
			"message": err.Error(),
		})
		return
	}

	// Delete advertisements
	if err := c.Container.GetService("advertisement").(base_services.InterfaceAdvertisementService).Delete(form.IDs); err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "failed to commit transaction",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete advertisement success"})
}

// GetOne retrieves a single advertisement
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
