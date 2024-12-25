package http_base_controller

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
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

type NoticeController struct {
	ctx     *gin.Context
	service base_services.InterfaceNoticeService
}

func NewNoticeController(
	ctx *gin.Context,
	service base_services.InterfaceNoticeService,
) InterfaceNoticeController {
	return &NoticeController{
		ctx:     ctx,
		service: service,
	}
}

func (c *NoticeController) Create() {
	var form struct {
		Title       string           `json:"title" binding:"required"`
		Description string           `json:"description"`
		Type        field.NoticeType `json:"type" binding:"required"`
		Status      field.Status     `json:"status" binding:"required"`
		StartTime   *time.Time       `json:"startTime" binding:"required"`
		EndTime     *time.Time       `json:"endTime" binding:"required"`
		IsPublic    bool             `json:"isPublic" binding:"required"`
		Path        string           `json:"path" binding:"required"`
		FileType    field.FileType   `json:"fileType"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{
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

	// 如果提供了 path，查找对应的文件
	var fileID *uint
	if form.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			c.ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID
	}

	notice := &base_models.Notice{
		Title:       form.Title,
		Description: form.Description,
		Type:        form.Type,
		Status:      status,
		StartTime:   startTime,
		EndTime:     endTime,
		IsPublic:    form.IsPublic,
		FileID:      fileID,
		FileType:    form.FileType,
	}

	if err := c.service.Create(notice); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create notice failed",
		})
		return
	}

	// 重新加载 notice 以获取关联的文件信息
	if err := databases.DB_CONN.Preload("File").First(notice, notice.ID).Error; err != nil {
		c.ctx.JSON(200, gin.H{
			"message": "create notice success, but failed to load file info",
			"data":    notice,
		})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "create notice success",
		"data":    notice,
	})
}

func (c *NoticeController) CreateMany() {
	var forms []struct {
		Title       string           `json:"title" binding:"required"`
		Description string           `json:"description"`
		Type        field.NoticeType `json:"type" binding:"required"`
		Status      field.Status     `json:"status" binding:"required"`
		StartTime   *time.Time       `json:"startTime" binding:"required"`
		EndTime     *time.Time       `json:"endTime" binding:"required"`
		IsPublic    bool             `json:"isPublic" binding:"required"`
		Path        string           `json:"path" binding:"required"`
		FileType    field.FileType   `json:"fileType"`
		Duration    int              `json:"duration" binding:"required"`
	}

	if err := c.ctx.ShouldBindJSON(&forms); err != nil {
		c.ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var notices []*base_models.Notice
	for _, form := range forms {
		notice := &base_models.Notice{
			Title:       form.Title,
			Description: form.Description,
			Type:        form.Type,
			Status:      form.Status,
			StartTime:   *form.StartTime,
			EndTime:     *form.EndTime,
			IsPublic:    form.IsPublic,
		}
		notices = append(notices, notice)
	}

	for _, notice := range notices {
		if err := c.service.Create(notice); err != nil {
			c.ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "create notice failed",
				"notice":  notice,
			})
			return
		}
	}

	c.ctx.JSON(200, gin.H{
		"message": "create notices success",
		"data":    notices,
	})
}

func (c *NoticeController) Get() {
	var searchQuery struct {
		Search string `form:"search"`
		Type   string `form:"type"`
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
		Desc:     true,
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

	notices, paginationResult, err := c.service.Get(queryMap, paginationMap)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": paginationResult,
	})
}

func (c *NoticeController) Update() {
	var form struct {
		ID          uint             `json:"id" binding:"required"`
		Title       *string          `json:"title"`
		Description *string          `json:"description"`
		Type        field.NoticeType `json:"type"`
		Status      field.Status     `json:"status"`
		StartTime   *time.Time       `json:"startTime"`
		EndTime     *time.Time       `json:"endTime"`
		IsPublic    *bool            `json:"isPublic"`
		FileID      *uint            `json:"fileId"`
		FileType    field.FileType   `json:"fileType"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get existing notice to check current values
	existingNotice, err := c.service.GetByID(form.ID)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Notice not found"})
		return
	}

	updates := map[string]interface{}{}
	if form.Title != nil {
		updates["title"] = *form.Title
	}
	if form.Description != nil {
		updates["description"] = *form.Description
	}
	if form.Type != "" {
		updates["type"] = form.Type
	}
	if form.Status != "" {
		updates["status"] = form.Status
	} else if existingNotice.Status == "" {
		updates["status"] = field.Status("active")
	}
	if form.StartTime != nil {
		updates["start_time"] = *form.StartTime
	} else if existingNotice.StartTime.IsZero() {
		updates["start_time"] = time.Now()
	}
	if form.EndTime != nil {
		updates["end_time"] = *form.EndTime
	} else if existingNotice.EndTime.IsZero() {
		updates["end_time"] = time.Now().AddDate(1, 0, 0)
	}
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}
	if form.FileID != nil {
		updates["file_id"] = *form.FileID
	}
	if form.FileType != "" {
		updates["file_type"] = form.FileType
	}

	notice, err := c.service.Update(form.ID, updates)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "update notice success",
		"data":    notice,
	})
}

func (c *NoticeController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Delete(form.IDs); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "delete notice success"})
}

func (c *NoticeController) GetOne() {
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	notice, err := c.service.GetByID(uint(id))
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": notice})
}
