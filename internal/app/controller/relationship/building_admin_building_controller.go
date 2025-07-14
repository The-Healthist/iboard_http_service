package http_relationship_controller

import (
	"strconv"

	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceBuildingAdminBuildingController interface {
	BindBuildings()
	UnbindBuildings()
	GetBuildingsByBuildingAdmin()
	GetBuildingAdminsByBuilding()
}

type BuildingAdminBuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminBuildingController(ctx *gin.Context, container *container.ServiceContainer) *BuildingAdminBuildingController {
	return &BuildingAdminBuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdminBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdminBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewBuildingAdminBuildingController(ctx, container)
		switch method {
		case "bindBuildings":
			controller.BindBuildings()
		case "unbindBuildings":
			controller.UnbindBuildings()
		case "getBuildingsByBuildingAdmin":
			controller.GetBuildingsByBuildingAdmin()
		case "getBuildingAdminsByBuilding":
			controller.GetBuildingAdminsByBuilding()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *BuildingAdminBuildingController) getService() relationship_service.InterfaceBuildingAdminBuildingService {
	return c.Container.GetService("buildingAdminBuilding").(relationship_service.InterfaceBuildingAdminBuildingService)
}

// 1. BindBuildings 绑定建筑管理员到建筑
// @Summary      绑定建筑管理员到建筑
// @Description  将一个或多个建筑管理员绑定到一个或多个建筑物
// @Tags         BuildingAdmin-Building
// @Accept       json
// @Produce      json
// @Param        bindInfo body object true "绑定信息"
// @Param        buildingAdminIds body []uint true "建筑管理员ID列表" example:"[1,2,3]"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "绑定结果信息，包含成功绑定的关系和已存在的绑定"
// @Failure      400  {object}  map[string]interface{} "输入参数错误"
// @Failure      404  {object}  map[string]interface{} "建筑管理员或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/building-admin-building/bind [post]
// @Security     BearerAuth
func (c *BuildingAdminBuildingController) BindBuildings() {
	var form struct {
		BuildingAdminIDs []uint `json:"buildingAdminIds" binding:"required,min=1"`
		BuildingIDs      []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	var response struct {
		Success               []map[string]interface{} `json:"success"`
		NotFoundAdmins        []uint                   `json:"notFoundBuildingAdmins,omitempty"`
		NotFoundBuildings     []uint                   `json:"notFoundBuildings,omitempty"`
		AlreadyBound          []map[string]interface{} `json:"alreadyBound,omitempty"`
		FailedBindingDetails  []map[string]interface{} `json:"failedBindingDetails,omitempty"`
		TotalSuccess          int                      `json:"totalSuccess"`
		TotalFailedBindings   int                      `json:"totalFailedBindings"`
		TotalAlreadyExisted   int                      `json:"totalAlreadyExisted"`
		TotalNotFoundRecords  int                      `json:"totalNotFoundRecords"`
		TotalProcessedRecords int                      `json:"totalProcessedRecords"`
	}

	service := c.getService()

	// 检查所有建筑管理员是否存在
	for _, adminID := range form.BuildingAdminIDs {
		exists, err := service.BuildingAdminExists(adminID)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			response.NotFoundAdmins = append(response.NotFoundAdmins, adminID)
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
	if len(response.NotFoundAdmins) > 0 || len(response.NotFoundBuildings) > 0 {
		response.TotalNotFoundRecords = len(response.NotFoundAdmins) + len(response.NotFoundBuildings)
		response.TotalProcessedRecords = len(form.BuildingAdminIDs) * len(form.BuildingIDs)
		c.Ctx.JSON(404, response)
		return
	}

	// 处理每个建筑管理员的绑定
	for _, adminID := range form.BuildingAdminIDs {
		// 获取当前绑定的建筑物列表
		currentBuildings, err := service.GetBuildingsByAdminID(adminID)
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
				"buildingAdminId":       adminID,
				"duplicateBuildingIds":  duplicateBindings,
				"alreadyBoundBuildings": len(duplicateBindings),
			})
			response.TotalAlreadyExisted += len(duplicateBindings)
		}

		// 执行有效的绑定
		if len(validBindings) > 0 {
			err := service.BindBuildings(adminID, validBindings)
			if err != nil {
				c.Ctx.JSON(400, gin.H{"error": "Failed to bind buildings: " + err.Error()})
				return
			}

			// Since the BindBuildings doesn't return failed bindings, we'll assume all were successful
			response.Success = append(response.Success, map[string]interface{}{
				"buildingAdminId":        adminID,
				"buildingIds":            validBindings,
				"totalSuccessfulRecords": len(validBindings),
			})
			response.TotalSuccess += len(validBindings)
		}
	}

	response.TotalProcessedRecords = len(form.BuildingAdminIDs) * len(form.BuildingIDs)
	c.Ctx.JSON(200, response)
}

// 2. UnbindBuildings 解绑建筑管理员与建筑
// @Summary      解绑建筑管理员与建筑
// @Description  解除一个建筑管理员与多个建筑物的绑定关系
// @Tags         BuildingAdmin-Building
// @Accept       json
// @Produce      json
// @Param        unbindInfo body object true "解绑信息"
// @Param        buildingAdminId body uint true "建筑管理员ID" example:"1"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误或建筑未绑定"
// @Failure      404  {object}  map[string]interface{} "建筑管理员或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/building-admin-building/unbind [post]
// @Security     BearerAuth
func (c *BuildingAdminBuildingController) UnbindBuildings() {
	var form struct {
		BuildingAdminID uint   `json:"buildingAdminId" binding:"required"`
		BuildingIDs     []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	service := c.getService()

	// 检查建筑管理员是否存在
	exists, err := service.BuildingAdminExists(form.BuildingAdminID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Building admin not found"})
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
	currentBuildings, err := service.GetBuildingsByAdminID(form.BuildingAdminID)
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
			"error":               "Some Buildings are not bound to the Building Admin",
			"notBoundBuildingIds": notBoundIDs,
		})
		return
	}

	if err := service.UnbindBuildings(form.BuildingAdminID, validUnbind); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Failed to unbind buildings: " + err.Error()})
		return
	}

	c.Ctx.JSON(200, map[string]interface{}{"message": "Buildings unbound successfully"})
}

// 3. GetBuildingsByBuildingAdmin 获取建筑管理员管理的建筑
// @Summary      获取建筑管理员管理的建筑
// @Description  根据建筑管理员ID获取其管理的所有建筑物信息
// @Tags         BuildingAdmin-Building
// @Accept       json
// @Produce      json
// @Param        buildingAdminId query int true "建筑管理员ID" example:"1"
// @Success      200  {object}  map[string]interface{} "建筑列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的建筑管理员ID"
// @Failure      404  {object}  map[string]interface{} "建筑管理员不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/building-admin-building/buildings [get]
// @Security     BearerAuth
func (c *BuildingAdminBuildingController) GetBuildingsByBuildingAdmin() {
	adminIDStr := c.Ctx.Query("buildingAdminId")
	if adminIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "buildingAdminId is required"})
		return
	}

	adminID, err := strconv.ParseUint(adminIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid buildingAdminId"})
		return
	}

	service := c.getService()

	// 检查建筑管理员是否存在
	exists, err := service.BuildingAdminExists(uint(adminID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Building admin not found"})
		return
	}

	buildings, err := service.GetBuildingsByAdminID(uint(adminID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch buildings"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": buildings})
}

// 4. GetBuildingAdminsByBuilding 获取建筑的管理员
// @Summary      获取建筑的管理员
// @Description  根据建筑ID获取管理该建筑的所有管理员信息
// @Tags         BuildingAdmin-Building
// @Accept       json
// @Produce      json
// @Param        buildingId query int true "建筑ID" example:"1"
// @Success      200  {object}  map[string]interface{} "建筑管理员列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的建筑ID"
// @Failure      404  {object}  map[string]interface{} "建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/building-admin-building/admins [get]
// @Security     BearerAuth
func (c *BuildingAdminBuildingController) GetBuildingAdminsByBuilding() {
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

	admins, err := service.GetAdminsByBuildingID(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch building admins"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": admins})
}
