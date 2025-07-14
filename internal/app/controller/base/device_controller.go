package http_base_controller

import (
	"fmt"
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceDeviceController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
	Login()
	GetDeviceAdvertisements()
	GetDeviceNotices()
	HealthTest()
}

type DeviceController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewDeviceController(ctx *gin.Context, container *container.ServiceContainer) *DeviceController {
	return &DeviceController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncDevice returns a gin.HandlerFunc for the specified method
func HandleFuncDevice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetOne()
		}
	case "login":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Login()
		}
	case "getDeviceAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceAdvertisements()
		}
	case "getDeviceNotices":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceNotices()
		}
	case "healthTest":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.HealthTest()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// 1.Create 创建设备
// @Summary      创建设备
// @Description  创建一个新的设备
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        device body object true "设备信息"
// @Param        deviceId formData string true "设备ID" example:"DEV1001"
// @Param        buildingId formData uint true "建筑ID" example:"1"
// @Param        settings formData object false "设备设置"
// @Param        settings.arrearageUpdateDuration formData int false "欠费更新间隔(秒)" example:"3600"
// @Param        settings.noticeUpdateDuration formData int false "通知更新间隔(秒)" example:"1800"
// @Param        settings.advertisementUpdateDuration formData int false "广告更新间隔(秒)" example:"3600"
// @Param        settings.advertisementPlayDuration formData int false "广告播放时长(秒)" example:"30"
// @Param        settings.noticePlayDuration formData int false "通知播放时长(秒)" example:"20"
// @Param        settings.spareDuration formData int false "空闲时长(秒)" example:"10"
// @Param        settings.noticeStayDuration formData int false "通知停留时长(秒)" example:"5"
// @Success      200  {object}  map[string]interface{} "返回创建的设备信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [post]
// @Security     BearerAuth
func (c *DeviceController) Create() {
	var form struct {
		DeviceID   string                     `json:"deviceId" binding:"required" example:"DEV1001"`
		BuildingID uint                       `json:"buildingId" binding:"required" example:"1"`
		Settings   base_models.DeviceSettings `json:"settings"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	device := &base_models.Device{
		DeviceID:   form.DeviceID,
		BuildingID: form.BuildingID,
		Settings:   form.Settings,
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Create(device); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create device failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create device success",
		"data":    device,
	})
}

// 2.Get 获取设备列表
// @Summary      获取设备列表
// @Description  根据查询条件获取设备列表，包含设备状态
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        search query string false "搜索关键词" example:"DEV"
// @Param        pageSize query int false "每页数量" default(10)
// @Param        pageNum query int false "页码" default(1)
// @Param        desc query bool false "是否降序" default(true)
// @Success      200  {object}  map[string]interface{} "返回设备列表和分页信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [get]
// @Security     BearerAuth
func (c *DeviceController) Get() {
	var searchQuery struct {
		Search string `form:"search" example:"DEV"`
	}
	if err := c.Ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pagination := struct {
		PageSize int  `form:"pageSize" example:"10"`
		PageNum  int  `form:"pageNum" example:"1"`
		Desc     bool `form:"desc" example:"true"`
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

	devices, paginationResult, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetWithStatus(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       devices,
		"pagination": paginationResult,
	})
}

// 3.Update 更新设备
// @Summary      更新设备
// @Description  更新设备信息，包括设备ID、所属建筑和设置
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        device body object true "设备更新信息"
// @Param        id formData uint true "设备ID" example:"1"
// @Param        deviceId formData string false "设备ID" example:"DEV1001-UPDATED"
// @Param        buildingId formData uint false "建筑ID" example:"2"
// @Param        settings formData object false "设备设置"
// @Param        settings.arrearageUpdateDuration formData int false "欠费更新间隔(秒)" example:"7200"
// @Param        settings.noticeUpdateDuration formData int false "通知更新间隔(秒)" example:"3600"
// @Param        settings.advertisementUpdateDuration formData int false "广告更新间隔(秒)" example:"7200"
// @Param        settings.advertisementPlayDuration formData int false "广告播放时长(秒)" example:"45"
// @Param        settings.noticePlayDuration formData int false "通知播放时长(秒)" example:"30"
// @Param        settings.spareDuration formData int false "空闲时长(秒)" example:"15"
// @Param        settings.noticeStayDuration formData int false "通知停留时长(秒)" example:"10"
// @Success      200  {object}  map[string]interface{} "返回更新后的设备信息，包含状态"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [put]
// @Security     BearerAuth
func (c *DeviceController) Update() {
	var form struct {
		ID         uint                        `json:"id" binding:"required" example:"1"`
		DeviceID   string                      `json:"deviceId" example:"DEV1001-UPDATED"`
		BuildingID uint                        `json:"buildingId" example:"2"`
		Settings   *base_models.DeviceSettings `json:"settings"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 如果需要更新 buildingId，先处理绑定关系
	if form.BuildingID != 0 {
		// 先解绑
		if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).UnbindDevice(form.ID); err != nil {
			c.Ctx.JSON(400, gin.H{"error": fmt.Sprintf("failed to unbind device: %v", err)})
			return
		}

		// 再绑定到新建筑物
		if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).BindDevices(form.BuildingID, []uint{form.ID}); err != nil {
			c.Ctx.JSON(400, gin.H{"error": fmt.Sprintf("failed to bind device: %v", err)})
			return
		}
	}

	// 更新其他字段
	updates := map[string]interface{}{}
	if form.DeviceID != "" {
		updates["device_id"] = form.DeviceID
	}
	if form.Settings != nil {
		updates["arrearage_update_duration"] = form.Settings.ArrearageUpdateDuration
		updates["notice_update_duration"] = form.Settings.NoticeUpdateDuration
		updates["advertisement_update_duration"] = form.Settings.AdvertisementUpdateDuration
		updates["advertisement_play_duration"] = form.Settings.AdvertisementPlayDuration
		updates["notice_play_duration"] = form.Settings.NoticePlayDuration
		updates["spare_duration"] = form.Settings.SpareDuration
		updates["notice_stay_duration"] = form.Settings.NoticeStayDuration
	}

	updatedDevice, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取设备状态
	deviceWithStatus := base_services.DeviceWithStatus{
		Device: *updatedDevice,
		Status: c.Container.GetService("device").(base_services.InterfaceDeviceService).CheckDeviceStatus(updatedDevice.ID),
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update device success",
		"data":    deviceWithStatus,
	})
}

// 4.Delete 删除设备
// @Summary      删除设备
// @Description  删除一个或多个设备
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        ids body object true "设备ID列表"
// @Param        ids.ids body []uint true "设备ID数组" example:"[1,2,3]"
// @Success      200  {object}  map[string]interface{} "删除成功消息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [delete]
// @Security     BearerAuth
func (c *DeviceController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required" example:"[1,2,3]"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete device success"})
}

// 5.GetOne 获取单个设备
// @Summary      获取单个设备
// @Description  根据ID获取设备详细信息
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        id path int true "设备ID" example:"1"
// @Success      200  {object}  map[string]interface{} "返回设备详细信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "设备不存在"
// @Router       /admin/device/{id} [get]
// @Security     BearerAuth
func (c *DeviceController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid device ID"})
		return
	}

	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByIDWithStatus(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get device success",
		"data":    device,
	})
}

// 6.CreateMany 批量创建设备
// @Summary      批量创建设备
// @Description  批量创建多个设备
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        devices body object true "设备信息数组"
// @Param        devices.devices body array true "设备数组"
// @Success      200  {object}  map[string]interface{} "返回创建的设备信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/devices [post]
// @Security     BearerAuth
func (c *DeviceController) CreateMany() {
	var form struct {
		Devices []struct {
			DeviceID   string                     `json:"deviceId" binding:"required" example:"DEV1001"`
			BuildingID uint                       `json:"buildingId" binding:"required" example:"1"`
			Settings   base_models.DeviceSettings `json:"settings"`
		} `json:"devices" binding:"required,dive"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	devices := make([]*base_models.Device, len(form.Devices))
	for i, d := range form.Devices {
		devices[i] = &base_models.Device{
			DeviceID:   d.DeviceID,
			BuildingID: d.BuildingID,
			Settings:   d.Settings,
		}
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).CreateMany(devices); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create devices failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create devices success",
		"data":    devices,
	})
}

// 7.Login 设备登录
// @Summary      设备登录
// @Description  设备登录获取JWT令牌
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        login body object true "登录信息"
// @Param        deviceId formData string true "设备ID" example:"DEV1001"
// @Success      200  {object}  map[string]interface{} "返回登录令牌和设备信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      401  {object}  map[string]interface{} "认证失败"
// @Router       /device/login [post]
// @Security     None
func (c *DeviceController) Login() {
	var form struct {
		DeviceID string `json:"deviceId" binding:"required" example:"DEV1001"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid form",
		})
		return
	}

	// Get device by deviceId
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid device ID",
		})
		return
	}

	// Generate JWT token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateDeviceToken(device)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to generate token",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Login success",
		"data":    device,
		"token":   token,
	})
}

// 8.GetDeviceAdvertisements 获取设备广告
// @Summary      获取设备广告
// @Description  根据设备ID获取分配给该设备的广告
// @Tags         Device
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{} "返回设备广告列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "未找到设备"
// @Router       /device/client/advertisements [get]
// @Security     JWT
func (c *DeviceController) GetDeviceAdvertisements() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
		return
	}

	advertisements, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceAdvertisements(deviceId)
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

// 9.GetDeviceNotices 获取设备通知
// @Summary      获取设备通知
// @Description  根据设备ID获取分配给该设备的通知
// @Tags         Device
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{} "返回设备通知列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Failure      404  {object}  map[string]interface{} "未找到设备"
// @Router       /device/client/notices [get]
// @Security     JWT
func (c *DeviceController) GetDeviceNotices() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
		return
	}

	notices, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceNotices(deviceId)
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

// 10.HealthTest 设备健康测试
// @Summary      设备健康测试
// @Description  设备上报健康状态，用于检测设备是否在线
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        data body object false "健康测试数据"
// @Success      200  {object}  map[string]interface{} "健康测试成功响应"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/client/health_test [post]
// @Security     JWT
func (c *DeviceController) HealthTest() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "No token claims found"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(401, gin.H{"error": "Invalid token claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(401, gin.H{"error": "Invalid device ID format"})
		return
	}

	// 获取设备信息
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 更新设备健康状态
	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).UpdateDeviceHealth(device.ID); err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "Failed to update device health status",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Health check successful",
	})
}
