package http_relationship_controller

import (
	"strconv"

	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceAdvertisementBuildingController interface {
	BindBuildings()
	UnbindBuildings()
	GetBuildingsByAdvertisement()
	GetAdvertisementsByBuilding()
}

type AdvertisementBuildingController struct {
	ctx     *gin.Context
	service relationship_service.InterfaceAdvertisementBuildingService
}

func NewAdvertisementBuildingController(
	ctx *gin.Context,
	service relationship_service.InterfaceAdvertisementBuildingService,
) InterfaceAdvertisementBuildingController {
	return &AdvertisementBuildingController{
		ctx:     ctx,
		service: service,
	}
}

func (c *AdvertisementBuildingController) BindBuildings() {
	var form struct {
		AdvertisementIDs []uint `json:"advertisementIds" binding:"required,min=1"`
		BuildingIDs      []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	var response struct {
		Success           []map[string]interface{} `json:"success"`
		NotFoundAds       []uint                   `json:"notFoundAdvertisements,omitempty"`
		NotFoundBuildings []uint                   `json:"notFoundBuildings,omitempty"`
		AlreadyBound      []map[string]interface{} `json:"alreadyBound,omitempty"`
	}

	// 检查所有广告是否存在
	for _, adID := range form.AdvertisementIDs {
		exists, err := c.service.AdvertisementExists(adID)
		if err != nil {
			c.ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			response.NotFoundAds = append(response.NotFoundAds, adID)
		}
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := c.service.BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		response.NotFoundBuildings = missingBuildings
	}

	// 如果有不存在的记录，直接返回错误
	if len(response.NotFoundAds) > 0 || len(response.NotFoundBuildings) > 0 {
		c.ctx.JSON(404, response)
		return
	}

	// 处理每个广告的绑定
	for _, adID := range form.AdvertisementIDs {
		// 获取当前绑定的建筑物列表
		currentBuildings, err := c.service.GetBuildingsByAdvertisementID(adID)
		if err != nil {
			c.ctx.JSON(500, gin.H{"error": "Failed to fetch current buildings"})
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
			if err := c.service.BindBuildings(adID, validBindings); err != nil {
				c.ctx.JSON(400, gin.H{"error": "Failed to bind buildings: " + err.Error()})
				return
			}
			response.Success = append(response.Success, map[string]interface{}{
				"advertisementId": adID,
				"buildingIds":     validBindings,
			})
		}
	}

	c.ctx.JSON(200, response)
}

func (c *AdvertisementBuildingController) UnbindBuildings() {
	var form struct {
		AdvertisementID uint   `json:"advertisementId" binding:"required"`
		BuildingIDs     []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	// 检查广告是否存在
	exists, err := c.service.AdvertisementExists(form.AdvertisementID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := c.service.BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		c.ctx.JSON(404, map[string]interface{}{
			"error":              "Some Buildings not found",
			"missingBuildingIds": missingBuildings,
		})
		return
	}

	// 获取当前绑定的建筑物列表
	currentBuildings, err := c.service.GetBuildingsByAdvertisementID(form.AdvertisementID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch current buildings"})
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
		c.ctx.JSON(400, map[string]interface{}{
			"error":               "Some Buildings are not bound to the Advertisement",
			"notBoundBuildingIds": notBoundIDs,
		})
		return
	}

	if err := c.service.UnbindBuildings(form.AdvertisementID, validUnbind); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Failed to unbind buildings: " + err.Error()})
		return
	}

	c.ctx.JSON(200, map[string]interface{}{"message": "Buildings unbound successfully"})
}

func (c *AdvertisementBuildingController) GetBuildingsByAdvertisement() {
	adIDStr := c.ctx.Query("advertisementId")
	if adIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "advertisementId is required"})
		return
	}

	adID, err := strconv.ParseUint(adIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisementId"})
		return
	}

	// 检查广告是否存在
	exists, err := c.service.AdvertisementExists(uint(adID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Advertisement not found"})
		return
	}

	buildings, err := c.service.GetBuildingsByAdvertisementID(uint(adID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch buildings"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": buildings})
}

func (c *AdvertisementBuildingController) GetAdvertisementsByBuilding() {
	buildingIDStr := c.ctx.Query("buildingId")
	if buildingIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "buildingId is required"})
		return
	}

	buildingID, err := strconv.ParseUint(buildingIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid buildingId"})
		return
	}

	// 检查建筑物是否存在
	exists, err := c.service.BuildingExists(uint(buildingID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Building not found"})
		return
	}

	advertisements, err := c.service.GetAdvertisementsByBuildingID(uint(buildingID))
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch advertisements"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": advertisements})
}
