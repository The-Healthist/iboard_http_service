package http_base_controller

import (
	"strconv"

	models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/gin-gonic/gin"
)

type InterfacePrinterController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
}

type PrinterController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewPrinterController(ctx *gin.Context, container *container.ServiceContainer) *PrinterController {
	return &PrinterController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncPrinter returns a gin.HandlerFunc for the specified method
func HandleFuncPrinter(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewPrinterController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewPrinterController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewPrinterController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewPrinterController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewPrinterController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Create 创建打印机
func (c *PrinterController) Create() {
	var form struct {
		DeviceID    *uint   `json:"deviceId"`
		DisplayName *string `json:"display_name"`
		IPAddress   *string `json:"ip_address"`
		Name        *string `json:"name"`
		State       *string `json:"state"`
		URI         *string `json:"uri"`
		Status      *string `json:"status"`
		Reason      *string `json:"reason"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	printer := &models.Printer{
		DeviceID:    form.DeviceID,
		DisplayName: form.DisplayName,
		IPAddress:   form.IPAddress,
		Name:        form.Name,
		State:       form.State,
		URI:         form.URI,
		Status:      form.Status,
		Reason:      form.Reason,
	}

	if err := c.Container.GetService("printer").(base_services.InterfacePrinterService).Create(printer); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create printer failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create printer success",
		"data":    printer,
	})
}

// Get 获取打印机列表
func (c *PrinterController) Get() {
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
		Desc:     true,
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

	printers, paginationResult, err := c.Container.GetService("printer").(base_services.InterfacePrinterService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       printers,
		"pagination": paginationResult,
	})
}

// Update 更新打印机
func (c *PrinterController) Update() {
	var form struct {
		ID          uint    `json:"id" binding:"required"`
		DeviceID    *uint   `json:"deviceId"`
		DisplayName *string `json:"display_name"`
		IPAddress   *string `json:"ip_address"`
		Name        *string `json:"name"`
		State       *string `json:"state"`
		URI         *string `json:"uri"`
		Status      *string `json:"status"`
		Reason      *string `json:"reason"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.DeviceID != nil {
		updates["device_id"] = form.DeviceID
	}
	if form.DisplayName != nil {
		updates["display_name"] = form.DisplayName
	}
	if form.IPAddress != nil {
		updates["ip_address"] = form.IPAddress
	}
	if form.Name != nil {
		updates["name"] = form.Name
	}
	if form.State != nil {
		updates["state"] = form.State
	}
	if form.URI != nil {
		updates["uri"] = form.URI
	}
	if form.Status != nil {
		updates["status"] = form.Status
	}
	if form.Reason != nil {
		updates["reason"] = form.Reason
	}

	updatedPrinter, err := c.Container.GetService("printer").(base_services.InterfacePrinterService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update printer success",
		"data":    updatedPrinter,
	})
}

// Delete 删除打印机
func (c *PrinterController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("printer").(base_services.InterfacePrinterService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete printer success"})
}

// GetOne 获取单个打印机
func (c *PrinterController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid printer ID"})
		return
	}

	printer, err := c.Container.GetService("printer").(base_services.InterfacePrinterService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get printer success",
		"data":    printer,
	})
}
