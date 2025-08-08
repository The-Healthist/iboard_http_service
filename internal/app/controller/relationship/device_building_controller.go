package http_relationship_controller

import (
	"strconv"

	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/gin-gonic/gin"
)

type DeviceBuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewDeviceBuildingController(ctx *gin.Context, container *container.ServiceContainer) *DeviceBuildingController {
	return &DeviceBuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

func HandleFuncDeviceBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewDeviceBuildingController(ctx, container)

		switch method {
		case "bindDevice":
			controller.BindDevice()
		case "unbindDevice":
			controller.UnbindDevice()
		case "getDevicesByBuilding":
			controller.GetDevicesByBuilding()
		case "getBuildingByDevice":
			controller.GetBuildingByDevice()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// 1. BindDevice 绑定设备到建筑
// @Summary      绑定设备到建筑
// @Description  将一个或多个设备绑定到一个建筑物
// @Tags         Device-Building
// @Accept       json
// @Produce      json
// @Param        bindInfo body object true "绑定信息"
// @Param        deviceIds body []uint true "设备ID列表" example:"[1,2,3]"
// @Param        buildingId body uint true "建筑ID" example:"4"
// @Success      200  {object}  map[string]interface{} "绑定成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误或绑定失败"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/device-building/bind [post]
// @Security     BearerAuth
func (c *DeviceBuildingController) BindDevice() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	var form struct {
		DeviceIDs  []uint `json:"deviceIds" binding:"required"`
		BuildingID uint   `json:"buildingId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		log.Warn("绑定设备表单无效 | %v | 错误: %v", requestID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("尝试绑定设备到建筑 | %v | 建筑ID: %d | 设备数量: %d", requestID, form.BuildingID, len(form.DeviceIDs))

	if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).BindDevices(form.BuildingID, form.DeviceIDs); err != nil {
		log.Error("绑定设备失败 | %v | 建筑ID: %d | 错误: %v", requestID, form.BuildingID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("绑定设备成功 | %v | 建筑ID: %d | 设备数量: %d", requestID, form.BuildingID, len(form.DeviceIDs))
	c.Ctx.JSON(200, gin.H{"message": "bind devices success"})
}

// 2. UnbindDevice 解绑设备与建筑
// @Summary      解绑设备与建筑
// @Description  解除一个设备与其关联的建筑物的绑定关系
// @Tags         Device-Building
// @Accept       json
// @Produce      json
// @Param        unbindInfo body object true "解绑信息"
// @Param        deviceId body uint true "设备ID" example:"1"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误或解绑失败"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/device-building/unbind [post]
// @Security     BearerAuth
func (c *DeviceBuildingController) UnbindDevice() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	var form struct {
		DeviceID uint `json:"deviceId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		log.Warn("解绑设备表单无效 | %v | 错误: %v", requestID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("尝试解绑设备 | %v | 设备ID: %d", requestID, form.DeviceID)

	if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).UnbindDevice(form.DeviceID); err != nil {
		log.Error("解绑设备失败 | %v | 设备ID: %d | 错误: %v", requestID, form.DeviceID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("解绑设备成功 | %v | 设备ID: %d", requestID, form.DeviceID)
	c.Ctx.JSON(200, gin.H{"message": "unbind device success"})
}

// 3. GetDevicesByBuilding 获取建筑绑定的设备
// @Summary      获取建筑绑定的设备
// @Description  获取与指定建筑物关联的所有设备信息及其状态
// @Tags         Device-Building
// @Accept       json
// @Produce      json
// @Param        buildingId query int true "建筑ID" example:"1"
// @Success      200  {object}  map[string]interface{} "设备列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的建筑ID或获取设备失败"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/device-building/devices [get]
// @Security     BearerAuth
func (c *DeviceBuildingController) GetDevicesByBuilding() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	buildingIDStr := c.Ctx.Query("buildingId")
	buildingID, err := strconv.ParseUint(buildingIDStr, 10, 64)
	if err != nil {
		log.Warn("获取建筑设备失败，无效的建筑ID | %v | 建筑ID: %s", requestID, buildingIDStr)
		c.Ctx.JSON(400, gin.H{"error": "Invalid building ID"})
		return
	}

	log.Info("获取建筑绑定的设备 | %v | 建筑ID: %d", requestID, buildingID)

	devices, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDevicesByBuildingWithStatus(uint(buildingID))
	if err != nil {
		log.Error("获取建筑设备失败 | %v | 建筑ID: %d | 错误: %v", requestID, buildingID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("获取建筑设备成功 | %v | 建筑ID: %d | 设备数量: %d", requestID, buildingID, len(devices))
	c.Ctx.JSON(200, gin.H{
		"message": "Get devices success",
		"data":    devices,
	})
}

// 4. GetBuildingByDevice 获取设备所属建筑
// @Summary      获取设备所属建筑
// @Description  根据设备ID获取其所属的建筑物信息
// @Tags         Device-Building
// @Accept       json
// @Produce      json
// @Param        deviceId query int true "设备ID" example:"1"
// @Success      200  {object}  map[string]interface{} "建筑信息"
// @Failure      400  {object}  map[string]interface{} "无效的设备ID或获取建筑失败"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/device-building/building [get]
// @Security     BearerAuth
func (c *DeviceBuildingController) GetBuildingByDevice() {
	// 获取请求ID
	requestID, _ := c.Ctx.Get(log.RequestIDKey)

	deviceIDStr := c.Ctx.Query("deviceId")
	deviceID, err := strconv.ParseUint(deviceIDStr, 10, 64)
	if err != nil {
		log.Warn("获取设备所属建筑失败，无效的设备ID | %v | 设备ID: %s", requestID, deviceIDStr)
		c.Ctx.JSON(400, gin.H{"error": "invalid device ID"})
		return
	}

	log.Info("获取设备所属建筑 | %v | 设备ID: %d", requestID, deviceID)

	building, err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).GetBuildingByDevice(uint(deviceID))
	if err != nil {
		log.Error("获取设备所属建筑失败 | %v | 设备ID: %d | 错误: %v", requestID, deviceID, err)
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	log.Info("获取设备所属建筑成功 | %v | 设备ID: %d | 建筑ID: %d", requestID, deviceID, building.ID)
	c.Ctx.JSON(200, gin.H{
		"data": building,
	})
}
