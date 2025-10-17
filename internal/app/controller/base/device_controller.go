package http_base_controller

import (
	"strconv"

	models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
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
	GetDeviceTopAdvertisements()
	GetDeviceFullAdvertisements()
	HealthTest()
	PrintersHealthCheck()
	PrintersCallback()
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
	case "getDeviceTopAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceTopAdvertisements()
		}
	case "getDeviceFullAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceFullAdvertisements()
		}
	case "healthTest":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.HealthTest()
		}
	case "printersHealthCheck":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.PrintersHealthCheck()
		}
	case "printersCallback":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.PrintersCallback()
		}
	case "getTopAdCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetTopAdCarousel() }
	case "updateTopAdCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).UpdateTopAdCarousel() }
	case "getFullAdCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetFullAdCarousel() }
	case "updateFullAdCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).UpdateFullAdCarousel() }
	case "getNoticeCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetNoticeCarousel() }
	case "updateNoticeCarousel":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).UpdateNoticeCarousel() }
	case "getTopAdCarouselResolved":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetTopAdCarouselResolved() }
	case "getFullAdCarouselResolved":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetFullAdCarouselResolved() }
	case "getNoticeCarouselResolved":
		return func(ctx *gin.Context) { NewDeviceController(ctx, container).GetNoticeCarouselResolved() }
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Create 创建设备
// @Summary      1. 创建设备
// @Description  创建一个新的显示设备，包含设备基本信息和设置参数
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId formData string true "设备ID" example:"DEV1001"
// @Param        buildingId formData int true "建筑ID" example:"1"
// @Param        settings.arrearageUpdateDuration formData int false "欠款更新间隔(分钟)" example:"5"
// @Param        settings.noticeUpdateDuration formData int false "通知更新间隔(分钟)" example:"10"
// @Param        settings.advertisementUpdateDuration formData int false "广告更新间隔(分钟)" example:"15"
// @Param        settings.appUpdateDuration formData int false "应用更新间隔(秒)" example:"600"
// @Param        settings.advertisementPlayDuration formData int false "广告播放时长(秒)" example:"30"
// @Param        settings.noticeStayDuration formData int false "通知停留时长(秒)" example:"5"
// @Param        settings.printPassWord formData string false "打印密码" example:"1090119"
// @Param        settings.bottomCarouselDuration formData int false "底部轮播切换时间(秒)" example:"10"
// @Param        settings.paymentTableOnePageDuration formData int false "缴费表格单页停留时间(秒)" example:"5"
// @Param        settings.normalToAnnouncementCarouselDuration formData int false "正常播放到公告轮播时间(秒)" example:"10"
// @Param        settings.announcementCarouselToFullAdsCarouselDuration formData int false "公告轮播到全屏广告轮播时间(秒)" example:"10"
// @Success      200  {object}  map[string]interface{} "返回创建的设备信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [post]
// @Security     BearerAuth
func (c *DeviceController) Create() {
	var form struct {
		DeviceID   string                `json:"deviceId" binding:"required" example:"DEV1001"`
		BuildingID uint                  `json:"buildingId" binding:"required" example:"1"`
		Settings   models.DeviceSettings `json:"settings"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	device := &models.Device{
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
// @Summary      2. 获取设备列表
// @Description  根据查询条件获取设备列表，包含设备状态和分页信息
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

// Update 更新设备
// @Summary      3. 更新设备
// @Description  更新设备信息，包括设备ID、所属建筑和设置参数
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        id formData int true "设备ID" example:"1"
// @Param        deviceId formData string false "设备ID" example:"DEV1001-UPDATED"
// @Param        buildingId formData int false "建筑ID" example:"2"
// @Param        settings.arrearageUpdateDuration formData int false "欠款更新间隔(分钟)" example:"5"
// @Param        settings.noticeUpdateDuration formData int false "通知更新间隔(分钟)" example:"10"
// @Param        settings.advertisementUpdateDuration formData int false "广告更新间隔(分钟)" example:"15"
// @Param        settings.appUpdateDuration formData int false "应用更新间隔(秒)" example:"600"
// @Param        settings.advertisementPlayDuration formData int false "广告播放时长(秒)" example:"30"
// @Param        settings.noticeStayDuration formData int false "通知停留时长(秒)" example:"5"
// @Param        settings.printPassWord formData string false "打印密码" example:"1090119"
// @Param        settings.bottomCarouselDuration formData int false "底部轮播切换时间(秒)" example:"10"
// @Param        settings.paymentTableOnePageDuration formData int false "缴费表格单页停留时间(秒)" example:"5"
// @Param        settings.normalToAnnouncementCarouselDuration formData int false "正常播放到公告轮播时间(秒)" example:"10"
// @Param        settings.announcementCarouselToFullAdsCarouselDuration formData int false "公告轮播到全屏广告轮播时间(秒)" example:"10"
// @Success      200  {object}  map[string]interface{} "返回更新后的设备信息"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /admin/device [put]
// @Security     BearerAuth
func (c *DeviceController) Update() {
	var form struct {
		ID         uint                   `json:"id" binding:"required" example:"1"`
		DeviceID   string                 `json:"deviceId" example:"DEV1001-UPDATED"`
		BuildingID uint                   `json:"buildingId" example:"2"`
		Settings   *models.DeviceSettings `json:"settings"`
		Status     string                 `json:"status"` // 忽略status字段，设备状态由服务器自动管理
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 创建更新映射
	updates := map[string]interface{}{}

	// 如果需要更新 buildingId，将其添加到更新映射中，让 device 服务处理
	if form.BuildingID != 0 {
		updates["buildingId"] = form.BuildingID
	}
	if form.DeviceID != "" {
		updates["device_id"] = form.DeviceID
	}
	if form.Settings != nil {
		// 防止时间设置为0，如果为0则使用默认值
		if form.Settings.ArrearageUpdateDuration > 0 {
			updates["arrearage_update_duration"] = form.Settings.ArrearageUpdateDuration
		} else if form.Settings.ArrearageUpdateDuration == 0 {
			updates["arrearage_update_duration"] = 5 // 默认值
		}

		if form.Settings.NoticeUpdateDuration > 0 {
			updates["notice_update_duration"] = form.Settings.NoticeUpdateDuration
		} else if form.Settings.NoticeUpdateDuration == 0 {
			updates["notice_update_duration"] = 10 // 默认值
		}

		if form.Settings.AdvertisementUpdateDuration > 0 {
			updates["advertisement_update_duration"] = form.Settings.AdvertisementUpdateDuration
		} else if form.Settings.AdvertisementUpdateDuration == 0 {
			updates["advertisement_update_duration"] = 15 // 默认值
		}

		if form.Settings.AppUpdateDuration > 0 {
			updates["app_update_duration"] = form.Settings.AppUpdateDuration
		} else if form.Settings.AppUpdateDuration == 0 {
			updates["app_update_duration"] = 600 // 默认值
		}

		if form.Settings.AdvertisementPlayDuration > 0 {
			updates["advertisement_play_duration"] = form.Settings.AdvertisementPlayDuration
		} else if form.Settings.AdvertisementPlayDuration == 0 {
			updates["advertisement_play_duration"] = 30 // 默认值
		}

		if form.Settings.PrintPassWord != "" {
			updates["print_pass_word"] = form.Settings.PrintPassWord
		}

		if form.Settings.NoticeStayDuration > 0 {
			updates["notice_stay_duration"] = form.Settings.NoticeStayDuration
		} else if form.Settings.NoticeStayDuration == 0 {
			updates["notice_stay_duration"] = 10 // 默认值
		}

		if form.Settings.BottomCarouselDuration > 0 {
			updates["bottom_carousel_duration"] = form.Settings.BottomCarouselDuration
		} else if form.Settings.BottomCarouselDuration == 0 {
			updates["bottom_carousel_duration"] = 10 // 默认值
		}

		if form.Settings.PaymentTableOnePageDuration > 0 {
			updates["payment_table_one_page_duration"] = form.Settings.PaymentTableOnePageDuration
		} else if form.Settings.PaymentTableOnePageDuration == 0 {
			updates["payment_table_one_page_duration"] = 5 // 默认值
		}

		if form.Settings.NormalToAnnouncementCarouselDuration > 0 {
			updates["normal_to_announcement_carousel_duration"] = form.Settings.NormalToAnnouncementCarouselDuration
		} else if form.Settings.NormalToAnnouncementCarouselDuration == 0 {
			updates["normal_to_announcement_carousel_duration"] = 10 // 默认值
		}

		if form.Settings.AnnouncementCarouselToFullAdsCarouselDuration > 0 {
			updates["announcement_carousel_to_full_ads_carousel_duration"] = form.Settings.AnnouncementCarouselToFullAdsCarouselDuration
		} else if form.Settings.AnnouncementCarouselToFullAdsCarouselDuration == 0 {
			updates["announcement_carousel_to_full_ads_carousel_duration"] = 10 // 默认值
		}
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
// @Summary      4. 删除设备
// @Description  删除一个或多个设备，支持批量删除操作
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
// @Summary      5. 获取单个设备
// @Description  根据ID获取设备详细信息，包含设备状态信息
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
// @Summary      6. 批量创建设备
// @Description  批量创建多个设备，支持一次性创建多个设备配置
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
			DeviceID   string                `json:"deviceId" binding:"required" example:"DEV1001"`
			BuildingID uint                  `json:"buildingId" binding:"required" example:"1"`
			Settings   models.DeviceSettings `json:"settings"`
		} `json:"devices" binding:"required,dive"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	devices := make([]*models.Device, len(form.Devices))
	for i, d := range form.Devices {
		devices[i] = &models.Device{
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
// @Summary      7. 设备登录
// @Description  设备登录获取JWT令牌，用于后续接口认证
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
// @Summary      8. 获取设备广告
// @Description  根据设备ID获取分配给该设备的广告列表
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
// @Summary      9. 获取设备通知
// @Description  根据设备ID获取分配给该设备的通知列表
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

// GetDeviceTopAdvertisements 获取设备顶部广告列表（包括top和topfull）
func (c *DeviceController) GetDeviceTopAdvertisements() {
	// Authenticate and extract deviceId
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
	ads, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceTopAdvertisements(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "Failed to get top advertisements"})
		return
	}
	c.Ctx.JSON(200, gin.H{"message": "Get top advertisements success", "data": ads})
}

// GetDeviceFullAdvertisements 获取设备全屏广告列表（包括full和topfull）
func (c *DeviceController) GetDeviceFullAdvertisements() {
	// Authenticate and extract deviceId
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
	ads, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceFullAdvertisements(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error(), "message": "Failed to get full advertisements"})
		return
	}
	c.Ctx.JSON(200, gin.H{"message": "Get full advertisements success", "data": ads})
}

// 10.HealthTest 设备健康测试
// @Summary      10. 设备健康测试
// @Description  设备上报健康状态，用于检测设备是否在线，更新设备最后活跃时间
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

// 11.GetTopAdCarousel 获取顶部广告轮播顺序
// @Summary      11. 获取顶部广告轮播顺序
// @Description  根据设备ID获取顶部广告轮播的排序ID列表
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回顶部广告轮播ID列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/top_advertisements [get]
// @Security     JWT
func (c *DeviceController) GetTopAdCarousel() {
	var form struct {
		DeviceID string `json:"deviceId" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	ids, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetTopAdCarousel(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": ids})
}

// 12.UpdateTopAdCarousel 更新顶部广告轮播顺序（全量替换）
// @Summary      12. 更新顶部广告轮播顺序
// @Description  更新设备顶部广告轮播顺序，支持全量替换现有顺序
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Param        data body []models.Advertisement true "广告数据数组"
// @Success      200  {object}  map[string]interface{} "返回更新后的完整广告列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/top_advertisements [put]
// @Security     JWT
func (c *DeviceController) UpdateTopAdCarousel() {
	var form struct {
		DeviceId string                 `json:"deviceId" binding:"required"`
		Data     []models.Advertisement `json:"data" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 提取 ID 列表
	var ids []uint
	for _, ad := range form.Data {
		ids = append(ids, ad.ID)
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).UpdateTopAdCarousel(device.ID, ids); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的完整列表
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetTopAdCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": list, "message": "Update top advertisements success"})
}

// 13.GetFullAdCarousel 获取全屏广告轮播顺序
// @Summary      13. 获取全屏广告轮播顺序
// @Description  根据设备ID获取全屏广告轮播的排序ID列表
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回全屏广告轮播ID列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/full_advertisements [get]
// @Security     JWT
func (c *DeviceController) GetFullAdCarousel() {
	var form struct {
		DeviceID string `json:"deviceId" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	ids, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetFullAdCarousel(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": ids})
}

// 14.UpdateFullAdCarousel 更新全屏广告轮播顺序（全量替换）
// @Summary      14. 更新全屏广告轮播顺序
// @Description  更新设备全屏广告轮播顺序，支持全量替换现有顺序
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Param        data body []models.Advertisement true "广告数据数组"
// @Success      200  {object}  map[string]interface{} "返回更新后的完整广告列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/full_advertisements [put]
// @Security     JWT
func (c *DeviceController) UpdateFullAdCarousel() {
	var form struct {
		DeviceId string                 `json:"deviceId" binding:"required"`
		Data     []models.Advertisement `json:"data" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 提取 ID 列表
	var ids []uint
	for _, ad := range form.Data {
		ids = append(ids, ad.ID)
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).UpdateFullAdCarousel(device.ID, ids); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的完整列表
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetFullAdCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": list, "message": "Update full advertisements success"})
}

// 15.GetNoticeCarousel 获取公告轮播顺序
// @Summary      15. 获取公告轮播顺序
// @Description  根据设备ID获取公告轮播的排序ID列表
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回公告轮播ID列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/notices [get]
// @Security     JWT
func (c *DeviceController) GetNoticeCarousel() {
	var form struct {
		DeviceID string `json:"deviceId" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	ids, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetNoticeCarousel(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": ids})
}

// 16.UpdateNoticeCarousel 更新公告轮播顺序（全量替换）
// @Summary      16. 更新公告轮播顺序
// @Description  更新设备公告轮播顺序，支持全量替换现有顺序
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Param        data body []models.Notice true "公告数据数组"
// @Success      200  {object}  map[string]interface{} "返回更新后的完整公告列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/notices [put]
// @Security     JWT
func (c *DeviceController) UpdateNoticeCarousel() {
	var form struct {
		DeviceId string          `json:"deviceId" binding:"required"`
		Data     []models.Notice `json:"data" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 根据设备ID字符串查找设备
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 提取 ID 列表
	var ids []uint
	for _, notice := range form.Data {
		ids = append(ids, notice.ID)
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).UpdateNoticeCarousel(device.ID, ids); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的完整列表
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetNoticeCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.Ctx.JSON(200, gin.H{"data": list, "message": "Update notices success"})
}

// 17.GetTopAdCarouselResolved 获取顶部广告详细列表 (管理员根据deviceID获取)
// @Summary      17. 获取顶部广告详细列表
// @Description  管理员根据设备ID获取顶部广告轮播的完整详细信息列表
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回顶部广告完整信息列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/top_advertisements [post]
// @Security     JWT
func (c *DeviceController) GetTopAdCarouselResolved() {
	// Determine deviceId: GET from token claims, POST from JSON body
	var deviceIdStr string
	if c.Ctx.Request.Method == "GET" {
		claimsVal, exists := c.Ctx.Get("claims")
		if !exists {
			c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		claimsMap, okClaims := claimsVal.(map[string]interface{})
		if !okClaims {
			c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
			return
		}
		deviceIdTmp, okDevice := claimsMap["deviceId"].(string)
		if !okDevice {
			c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
			return
		}
		deviceIdStr = deviceIdTmp
	} else {
		var form struct {
			DeviceId string `json:"deviceId" binding:"required"`
		}
		if err := c.Ctx.ShouldBindJSON(&form); err != nil || form.DeviceId == "" {
			c.Ctx.JSON(400, gin.H{"error": "Device ID is required"})
			return
		}
		deviceIdStr = form.DeviceId
	}
	// Fetch device
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(deviceIdStr)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 获取该设备的顶部广告轮播列表（返回完整对象）
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetTopAdCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": list, "message": "Get top advertisements success"})
}

// 18.GetFullAdCarouselResolved 获取全屏广告详细列表 (管理员根据deviceID获取)
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回全屏广告完整信息列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/full_advertisements [post]
// @Security     JWT
func (c *DeviceController) GetFullAdCarouselResolved() {
	// Determine deviceId: GET from token claims, POST from JSON
	var deviceIdStr string
	if c.Ctx.Request.Method == "GET" {
		claimsVal, exists := c.Ctx.Get("claims")
		if !exists {
			c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		claimsMap, ok := claimsVal.(map[string]interface{})
		if !ok {
			c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
			return
		}
		tmp, ok2 := claimsMap["deviceId"].(string)
		if !ok2 {
			c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
			return
		}
		deviceIdStr = tmp
	} else {
		var form struct {
			DeviceId string `json:"deviceId" binding:"required"`
		}
		if err := c.Ctx.ShouldBindJSON(&form); err != nil || form.DeviceId == "" {
			c.Ctx.JSON(400, gin.H{"error": "Device ID is required"})
			return
		}
		deviceIdStr = form.DeviceId
	}
	// Fetch device by deviceId
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(deviceIdStr)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 获取该设备的全屏广告轮播列表（返回完整对象）
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetFullAdCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": list, "message": "Get full advertisements success"})
}

// 19.GetNoticeCarouselResolved 获取公告详细列表 (管理员根据deviceID获取)
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        deviceId body string true "设备ID" example:"DEVICE_1DA24A3A"
// @Success      200  {object}  map[string]interface{} "返回公告完整信息列表"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/carousel/notices [post]
// @Security     JWT
func (c *DeviceController) GetNoticeCarouselResolved() {
	// Determine deviceId: GET from token claims for device client, POST JSON for admin
	var deviceIdStr string
	if c.Ctx.Request.Method == "GET" {
		claimsVal, exists := c.Ctx.Get("claims")
		if !exists {
			c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		claimsMap, ok := claimsVal.(map[string]interface{})
		if !ok {
			c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
			return
		}
		tmp, ok2 := claimsMap["deviceId"].(string)
		if !ok2 {
			c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
			return
		}
		deviceIdStr = tmp
	} else {
		var form struct {
			DeviceId string `json:"deviceId" binding:"required"`
		}
		if err := c.Ctx.ShouldBindJSON(&form); err != nil || form.DeviceId == "" {
			c.Ctx.JSON(400, gin.H{"error": "Device ID is required"})
			return
		}
		deviceIdStr = form.DeviceId
	}
	// Fetch device
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(deviceIdStr)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 获取该设备的公告轮播列表（返回完整对象）
	list, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetNoticeCarouselResolved(device.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": list, "message": "Get notices success"})
}

// PrintersHealthCheck 打印机健康检查接口
// @Summary      打印机健康检查
// @Description  设备客户端上报香橙派服务状态和打印机列表，系统会智能同步打印机数据
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        data body object true "打印机健康检查数据"
// @Success      200  {object}  map[string]interface{} "健康检查成功响应"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/client/printers/health [post]
// @Security     JWT
func (c *DeviceController) PrintersHealthCheck() {
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

	// 解析请求数据
	var form struct {
		OrangePi struct {
			IP           *string `json:"ip"`
			Port         *int    `json:"port"`
			Status       string  `json:"status" binding:"required"` // online, offline, not_configured
			ResponseTime *int    `json:"response_time"`
			Reason       *string `json:"reason"`
			ErrorCode    *string `json:"error_code"`
		} `json:"orange_pi" binding:"required"`
		Printers []struct {
			DisplayName *string `json:"display_name"`
			IPAddress   *string `json:"ip_address"`
			Name        *string `json:"name"`
			State       *string `json:"state"`
			URI         *string `json:"uri"`
			Status      *string `json:"status"`
			Reason      *string `json:"reason"`
		} `json:"printers"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid request format",
		})
		return
	}

	// 验证 orange_pi.status 的值
	validStatuses := map[string]bool{"online": true, "offline": true, "not_configured": true}
	if !validStatuses[form.OrangePi.Status] {
		c.Ctx.JSON(400, gin.H{
			"error":   "invalid_status",
			"message": "orange_pi.status must be one of: online, offline, not_configured",
		})
		return
	}

	// 转换打印机数据为 interface{} 切片
	printersInterface := make([]interface{}, len(form.Printers))
	for i, p := range form.Printers {
		printerMap := make(map[string]interface{})

		if p.DisplayName != nil {
			printerMap["display_name"] = *p.DisplayName
		}
		if p.IPAddress != nil {
			printerMap["ip_address"] = *p.IPAddress
		}
		if p.Name != nil {
			printerMap["name"] = *p.Name
		}
		if p.State != nil {
			printerMap["state"] = *p.State
		}
		if p.URI != nil {
			printerMap["uri"] = *p.URI
		}
		if p.Status != nil {
			printerMap["status"] = *p.Status
		}
		if p.Reason != nil {
			printerMap["reason"] = *p.Reason
		}

		printersInterface[i] = printerMap
	}

	// 调用 service 处理健康检查
	result, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).HandlePrintersHealthCheck(
		device.ID,
		form.OrangePi.IP,
		form.OrangePi.Port,
		form.OrangePi.Status,
		form.OrangePi.ResponseTime,
		form.OrangePi.Reason,
		form.OrangePi.ErrorCode,
		printersInterface,
	)

	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to process health check",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, result)
}

// PrintersCallback 打印回调接口
// @Summary      打印回调
// @Description  设备客户端上报打印结果，包括成功或失败原因
// @Tags         Device
// @Accept       json
// @Produce      json
// @Param        data body object true "打印回调数据"
// @Success      200  {object}  map[string]interface{} "打印回调成功响应"
// @Failure      400  {object}  map[string]interface{} "错误信息"
// @Router       /device/client/printers/callback [post]
// @Security     JWT
func (c *DeviceController) PrintersCallback() {
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

	// 解析请求数据
	var form struct {
		Printers []struct {
			DisplayName *string `json:"display_name"`
			IPAddress   *string `json:"ip_address"`
			Name        *string `json:"name"`
			State       *string `json:"state"`
			URI         *string `json:"uri"`
			Status      *string `json:"status"` // print status: "online" (成功) or "offline" (失败)
			Reason      *string `json:"reason"` // 如果打印失败，这里包含失败原因
		} `json:"printers"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid request format",
		})
		return
	}

	// 转换为 Printer models
	printers := make([]models.Printer, len(form.Printers))
	for i, p := range form.Printers {
		printers[i] = models.Printer{
			DisplayName: p.DisplayName,
			IPAddress:   p.IPAddress,
			Name:        p.Name,
			State:       p.State,
			URI:         p.URI,
			Status:      p.Status,
			Reason:      p.Reason,
		}
	}

	// 批量更新或创建打印机（记录打印结果）
	if err := c.Container.GetService("printer").(base_services.InterfacePrinterService).BatchUpdateOrCreate(device.ID, printers); err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "Failed to update printers callback",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"success": true,
		"message": "Printers callback processed",
		"count":   len(printers),
	})
}
