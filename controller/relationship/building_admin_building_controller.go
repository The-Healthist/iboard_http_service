package http_relationship_controller

import (
	"strconv"

	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceBuildingAdminBuildingController interface {
	BindBuildings()
	UnbindBuildings()
	GetBuildingsByAdmin()
	GetAdminsByBuilding()
}

type BuildingAdminBuildingController struct {
	ctx     *gin.Context
	service relationship_service.InterfaceBuildingAdminBuildingService
}

func NewBuildingAdminBuildingController(
	ctx *gin.Context,
	service relationship_service.InterfaceBuildingAdminBuildingService,
) InterfaceBuildingAdminBuildingController {
	return &BuildingAdminBuildingController{
		ctx:     ctx,
		service: service,
	}
}

func (c *BuildingAdminBuildingController) BindBuildings() {
	var form struct {
		BuildingAdminIDs []uint `json:"buildingAdminIds" binding:"required,min=1"`
		BuildingIDs      []uint `json:"buildingIds" binding:"required,min=1"`
	}

	// 绑定 JSON 参数并验证
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	var response struct {
		Success           []map[string]interface{} `json:"success"`
		NotFoundAdmins    []uint                   `json:"notFoundBuildingAdmins,omitempty"`
		NotFoundBuildings []uint                   `json:"notFoundBuildings,omitempty"`
		AlreadyBound      []map[string]interface{} `json:"alreadyBound,omitempty"`
	}

	// 检查所有 BuildingAdmin 是否存在
	for _, adminID := range form.BuildingAdminIDs {
		exists, err := c.service.BuildingAdminExists(adminID)
		if err != nil {
			c.ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			response.NotFoundAdmins = append(response.NotFoundAdmins, adminID)
		}
	}

	// 检查所有 Building 是否存在
	missingBuildings, err := c.service.BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		response.NotFoundBuildings = missingBuildings
	}

	// 如果有不存在的记录，直接返回错误
	if len(response.NotFoundAdmins) > 0 || len(response.NotFoundBuildings) > 0 {
		c.ctx.JSON(404, response)
		return
	}

	// 处理每个管理员的绑定
	for _, adminID := range form.BuildingAdminIDs {
		// 获取当前绑定的建筑物列表
		currentBuildings, err := c.service.GetBuildingsByAdminID(adminID)
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
				"buildingAdminId":      adminID,
				"duplicateBuildingIds": duplicateBindings,
			})
		}

		// 执行有效的绑定
		if len(validBindings) > 0 {
			if err := c.service.BindBuildings(adminID, validBindings); err != nil {
				c.ctx.JSON(400, gin.H{"error": "Failed to bind buildings: " + err.Error()})
				return
			}
			response.Success = append(response.Success, map[string]interface{}{
				"buildingAdminId": adminID,
				"buildingIds":     validBindings,
			})
		}
	}

	c.ctx.JSON(200, response)
}

func (c *BuildingAdminBuildingController) UnbindBuildings() {
	var form struct {
		BuildingAdminID uint   `json:"buildingAdminId" binding:"required"`
		BuildingIDs     []uint `json:"buildingIds" binding:"required,min=1"`
	}

	// 绑定 JSON 参数并验证
	if err := c.ctx.ShouldBindJSON(&form); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	// 检查 BuildingAdmin 是否存在
	exists, err := c.service.BuildingAdminExists(form.BuildingAdminID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "BuildingAdmin not found"})
		return
	}

	// 检查所有 Building 是否存在
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
	currentBuildings, err := c.service.GetBuildingsByAdminID(form.BuildingAdminID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch current buildings"})
		return
	}

	// 构建已绑定建筑物的 ID 集合
	currentBoundMap := make(map[uint]bool)
	for _, b := range currentBuildings {
		currentBoundMap[b.ID] = true
	}

	// 检查是否有未绑定的请求
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
			"error":               "Some Buildings are not bound to the BuildingAdmin",
			"notBoundBuildingIds": notBoundIDs,
		})
		return
	}

	// 解绑建筑物
	if err := c.service.UnbindBuildings(form.BuildingAdminID, validUnbind); err != nil {
		c.ctx.JSON(400, gin.H{"error": "Failed to unbind buildings: " + err.Error()})
		return
	}

	c.ctx.JSON(200, map[string]interface{}{"message": "Buildings unbound successfully"})
}

func (c *BuildingAdminBuildingController) GetBuildingsByAdmin() {
	buildingAdminIDStr := c.ctx.Query("buildingAdminId")
	if buildingAdminIDStr == "" {
		c.ctx.JSON(400, gin.H{"error": "buildingAdminId is required"})
		return
	}

	buildingAdminID, err := strconv.ParseUint(buildingAdminIDStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid buildingAdminId"})
		return
	}

	adminID := uint(buildingAdminID)

	// 检查 BuildingAdmin 是否存在
	exists, err := c.service.BuildingAdminExists(adminID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "BuildingAdmin not found"})
		return
	}

	buildings, err := c.service.GetBuildingsByAdminID(adminID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch buildings"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": buildings})
}

func (c *BuildingAdminBuildingController) GetAdminsByBuilding() {
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

	bID := uint(buildingID)

	// 检查 Building 是否存在
	exists, err := c.service.BuildingExists(bID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.ctx.JSON(404, gin.H{"error": "Building not found"})
		return
	}

	admins, err := c.service.GetAdminsByBuildingID(bID)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": "Failed to fetch administrators"})
		return
	}

	c.ctx.JSON(200, gin.H{"data": admins})
}
