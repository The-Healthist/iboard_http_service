package http_relationship_controller

import (
	"strconv"

	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementBuildingController interface {
	BindBuildings()
	UnbindBuildings()
	GetBuildingsByAdvertisement()
	GetAdvertisementsByBuilding()
}

type AdvertisementBuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewAdvertisementBuildingController(ctx *gin.Context, container *container.ServiceContainer) *AdvertisementBuildingController {
	return &AdvertisementBuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncAdvertisementBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncAdvertisementBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewAdvertisementBuildingController(ctx, container)
		switch method {
		case "bindBuildings":
			controller.BindBuildings()
		case "unbindBuildings":
			controller.UnbindBuildings()
		case "getBuildingsByAdvertisement":
			controller.GetBuildingsByAdvertisement()
		case "getAdvertisementsByBuilding":
			controller.GetAdvertisementsByBuilding()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *AdvertisementBuildingController) getService() relationship_service.InterfaceAdvertisementBuildingService {
	return c.Container.GetService("advertisementBuilding").(relationship_service.InterfaceAdvertisementBuildingService)
}

// BindBuildings 将广告绑定到建筑
// @Summary      绑定广告到建筑
// @Description  将一个或多个广告绑定到一个或多个建筑物
// @Tags         Advertisement-Building
// @Accept       json
// @Produce      json
// @Param        bindingInfo body object true "绑定信息"
// @Param        advertisementIds body []uint true "广告ID列表" example:"[1,2,3]"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "绑定结果信息，包含成功绑定的关系和已存在的绑定"
// @Failure      400  {object}  map[string]interface{} "输入参数错误"
// @Failure      404  {object}  map[string]interface{} "广告或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/advertisement-building/bind [post]
// @Security     BearerAuth
func (c *AdvertisementBuildingController) BindBuildings() {
	var form struct {
		AdvertisementIDs []uint `json:"advertisementIds" binding:"required,min=1"`
		BuildingIDs      []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	var response struct {
		Success           []map[string]interface{} `json:"success"`
		NotFoundAds       []uint                   `json:"notFoundAdvertisements,omitempty"`
		NotFoundBuildings []uint                   `json:"notFoundBuildings,omitempty"`
		AlreadyBound      []map[string]interface{} `json:"alreadyBound,omitempty"`
	}

	service := c.getService()

	// 检查所有广告是否存在
	for _, adID := range form.AdvertisementIDs {
		exists, err := service.AdvertisementExists(adID)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			response.NotFoundAds = append(response.NotFoundAds, adID)
		}
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := service.BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		response.NotFoundBuildings = missingBuildings
	}

	// 如果有不存在的记录，直接返回错误
	if len(response.NotFoundAds) > 0 || len(response.NotFoundBuildings) > 0 {
		c.Ctx.JSON(404, response)
		return
	}

	// 处理每个广告的绑定
	for _, adID := range form.AdvertisementIDs {
		// 获取当前绑定的建筑物列表
		currentBuildings, err := service.GetBuildingsByAdvertisementID(adID)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "Failed to fetch current buildings"})
			return
		}

		// 检查重复绑定
		alreadyBoundMap := make(map[uint]bool)
		for _, b := range currentBuildings {
			alreadyBoundMap[b.ID] = true
		}

		var duplicateBindings []uint
		var validBindings []uint
		for _, id := range form.BuildingIDs {
			if alreadyBoundMap[id] {
				duplicateBindings = append(duplicateBindings, id)
			} else {
				validBindings = append(validBindings, id)
			}
		}

		// 记录已经绑定的关系
		if len(duplicateBindings) > 0 {
			response.AlreadyBound = append(response.AlreadyBound, map[string]interface{}{
				"advertisementId":      adID,
				"duplicateBuildingIds": duplicateBindings,
			})
		}

		// 执行有效的绑定
		if len(validBindings) > 0 {
			if err := service.BindBuildings(adID, validBindings); err != nil {
				c.Ctx.JSON(400, gin.H{"error": "Failed to bind buildings: " + err.Error()})
				return
			}
			response.Success = append(response.Success, map[string]interface{}{
				"advertisementId": adID,
				"buildingIds":     validBindings,
			})
		}
	}

	c.Ctx.JSON(200, response)
}

// UnbindBuildings 解除广告与建筑的绑定
// @Summary      解绑广告与建筑
// @Description  解除一个广告与多个建筑物的绑定关系
// @Tags         Advertisement-Building
// @Accept       json
// @Produce      json
// @Param        unbindInfo body object true "解绑信息"
// @Param        advertisementId body uint true "广告ID" example:"1"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误或建筑未绑定"
// @Failure      404  {object}  map[string]interface{} "广告或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/advertisement-building/unbind [post]
// @Security     BearerAuth
func (c *AdvertisementBuildingController) UnbindBuildings() {
	var form struct {
		AdvertisementID uint   `json:"advertisementId" binding:"required"`
		BuildingIDs     []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	service := c.getService()

	// 检查广告是否存在
	exists, err := service.AdvertisementExists(form.AdvertisementID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := service.BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		c.Ctx.JSON(404, map[string]interface{}{
			"error":              "Some Buildings not found",
			"missingBuildingIds": missingBuildings,
		})
		return
	}

	// 获取当前绑定的建筑物列表
	currentBuildings, err := service.GetBuildingsByAdvertisementID(form.AdvertisementID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch current buildings"})
		return
	}

	// 检查未绑定的请求
	currentBoundMap := make(map[uint]bool)
	for _, b := range currentBuildings {
		currentBoundMap[b.ID] = true
	}

	var notBoundIDs []uint
	var validUnbind []uint
	for _, id := range form.BuildingIDs {
		if !currentBoundMap[id] {
			notBoundIDs = append(notBoundIDs, id)
		} else {
			validUnbind = append(validUnbind, id)
		}
	}

	if len(notBoundIDs) > 0 {
		c.Ctx.JSON(400, map[string]interface{}{
			"error":               "Some Buildings are not bound to the Advertisement",
			"notBoundBuildingIds": notBoundIDs,
		})
		return
	}

	if err := service.UnbindBuildings(form.AdvertisementID, validUnbind); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Failed to unbind buildings: " + err.Error()})
		return
	}

	c.Ctx.JSON(200, map[string]interface{}{"message": "Buildings unbound successfully"})
}

// GetBuildingsByAdvertisement 获取广告绑定的建筑
// @Summary      获取广告绑定的建筑
// @Description  根据广告ID获取所有与之绑定的建筑物信息
// @Tags         Advertisement-Building
// @Accept       json
// @Produce      json
// @Param        advertisementId query int true "广告ID" example:"1"
// @Success      200  {object}  map[string]interface{} "建筑列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的广告ID"
// @Failure      404  {object}  map[string]interface{} "广告不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/advertisement-building/buildings [get]
// @Security     BearerAuth
func (c *AdvertisementBuildingController) GetBuildingsByAdvertisement() {
	adIDStr := c.Ctx.Query("advertisementId")
	if adIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	adID, err := strconv.ParseUint(adIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	service := c.getService()

	// 检查广告是否存在
	exists, err := service.AdvertisementExists(uint(adID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	buildings, err := service.GetBuildingsByAdvertisementID(uint(adID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch buildings"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": buildings})
}

// GetAdvertisementsByBuilding 获取建筑绑定的广告
// @Summary      获取建筑绑定的广告
// @Description  根据建筑ID获取所有与之绑定的广告信息
// @Tags         Advertisement-Building
// @Accept       json
// @Produce      json
// @Param        buildingId query int true "建筑ID" example:"1"
// @Success      200  {object}  map[string]interface{} "广告列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的建筑ID"
// @Failure      404  {object}  map[string]interface{} "建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/advertisement-building/advertisements [get]
// @Security     BearerAuth
func (c *AdvertisementBuildingController) GetAdvertisementsByBuilding() {
	buildingIDStr := c.Ctx.Query("buildingId")
	if buildingIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "buildingId is required"})
		return
	}

	buildingID, err := strconv.ParseUint(buildingIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid buildingId"})
		return
	}

	service := c.getService()

	// 检查建筑物是否存在
	exists, err := service.BuildingExists(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Building not found"})
		return
	}

	advertisements, err := service.GetAdvertisementsByBuildingID(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch advertisements"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": advertisements})
}
