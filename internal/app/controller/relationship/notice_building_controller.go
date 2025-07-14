package http_relationship_controller

import (
	"strconv"

	container "github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/internal/domain/services/relationship"
	"github.com/gin-gonic/gin"
)

type InterfaceNoticeBuildingController interface {
	BindBuildings()
	UnbindBuildings()
	GetBuildingsByNotice()
	GetNoticesByBuilding()
}

type NoticeBuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewNoticeBuildingController(ctx *gin.Context, container *container.ServiceContainer) *NoticeBuildingController {
	return &NoticeBuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncNoticeBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncNoticeBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := NewNoticeBuildingController(ctx, container)
		switch method {
		case "bindBuildings":
			controller.BindBuildings()
		case "unbindBuildings":
			controller.UnbindBuildings()
		case "getBuildingsByNotice":
			controller.GetBuildingsByNotice()
		case "getNoticesByBuilding":
			controller.GetNoticesByBuilding()
		default:
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *NoticeBuildingController) getService() relationship_service.InterfaceNoticeBuildingService {
	return c.Container.GetService("noticeBuilding").(relationship_service.InterfaceNoticeBuildingService)
}

// 1. BindBuildings 绑定通知到建筑
// @Summary      绑定通知到建筑
// @Description  将一个或多个通知绑定到一个或多个建筑物
// @Tags         Notice-Building
// @Accept       json
// @Produce      json
// @Param        bindingInfo body object true "绑定信息"
// @Param        noticeIds body []uint true "通知ID列表" example:"[1,2,3]"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "绑定结果信息，包含成功绑定的关系和已存在的绑定"
// @Failure      400  {object}  map[string]interface{} "输入参数错误"
// @Failure      404  {object}  map[string]interface{} "通知或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/notice-building/bind [post]
// @Security     BearerAuth
func (c *NoticeBuildingController) BindBuildings() {
	var form struct {
		NoticeIDs   []uint `json:"noticeIds"`
		BuildingIDs []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	var response struct {
		Success           []map[string]interface{} `json:"success"`
		NotFoundNotices   []uint                   `json:"notFoundNotices,omitempty"`
		NotFoundBuildings []uint                   `json:"notFoundBuildings,omitempty"`
		AlreadyBound      []map[string]interface{} `json:"alreadyBound,omitempty"`
	}

	// If noticeIds is empty, return success with empty results
	if len(form.NoticeIDs) == 0 {
		c.Ctx.JSON(200, response)
		return
	}

	// 检查所有通知是否存在
	for _, noticeID := range form.NoticeIDs {
		exists, err := c.getService().NoticeExists(noticeID)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			response.NotFoundNotices = append(response.NotFoundNotices, noticeID)
		}
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := c.getService().BulkCheckBuildingsExistence(form.BuildingIDs)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if len(missingBuildings) > 0 {
		response.NotFoundBuildings = missingBuildings
	}

	// 如果有不存在的记录，直接返回错误
	if len(response.NotFoundNotices) > 0 || len(response.NotFoundBuildings) > 0 {
		c.Ctx.JSON(404, response)
		return
	}

	// 处理每个通知的绑定
	for _, noticeID := range form.NoticeIDs {
		// 获取当前绑定的建筑物列表
		currentBuildings, err := c.getService().GetBuildingsByNoticeID(noticeID)
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
				"noticeId":             noticeID,
				"duplicateBuildingIds": duplicateBindings,
			})
		}

		// 执行有效的绑定
		if len(validBindings) > 0 {
			if err := c.getService().BindBuildings(noticeID, validBindings); err != nil {
				c.Ctx.JSON(400, gin.H{"error": "Failed to bind buildings: " + err.Error()})
				return
			}
			response.Success = append(response.Success, map[string]interface{}{
				"noticeId":    noticeID,
				"buildingIds": validBindings,
			})
		}
	}

	c.Ctx.JSON(200, response)
}

// 2. UnbindBuildings 解绑通知与建筑
// @Summary      解绑通知与建筑
// @Description  解除一个通知与多个建筑物的绑定关系
// @Tags         Notice-Building
// @Accept       json
// @Produce      json
// @Param        unbindInfo body object true "解绑信息"
// @Param        noticeId body uint true "通知ID" example:"1"
// @Param        buildingIds body []uint true "建筑ID列表" example:"[4,5,6]"
// @Success      200  {object}  map[string]interface{} "解绑成功消息"
// @Failure      400  {object}  map[string]interface{} "输入参数错误或建筑未绑定"
// @Failure      404  {object}  map[string]interface{} "通知或建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/notice-building/unbind [post]
// @Security     BearerAuth
func (c *NoticeBuildingController) UnbindBuildings() {
	var form struct {
		NoticeID    uint   `json:"noticeId" binding:"required"`
		BuildingIDs []uint `json:"buildingIds" binding:"required,min=1"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid input parameters: " + err.Error()})
		return
	}

	// 检查通知是否存在
	exists, err := c.getService().NoticeExists(form.NoticeID)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Notice not found"})
		return
	}

	// 检查所有建筑物是否存在
	missingBuildings, err := c.getService().BulkCheckBuildingsExistence(form.BuildingIDs)
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
	currentBuildings, err := c.getService().GetBuildingsByNoticeID(form.NoticeID)
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
			"error":               "Some Buildings are not bound to the Notice",
			"notBoundBuildingIds": notBoundIDs,
		})
		return
	}

	if err := c.getService().UnbindBuildings(form.NoticeID, validUnbind); err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Failed to unbind buildings: " + err.Error()})
		return
	}

	c.Ctx.JSON(200, map[string]interface{}{"message": "Buildings unbound successfully"})
}

// 3. GetBuildingsByNotice 获取通知绑定的建筑
// @Summary      获取通知绑定的建筑
// @Description  根据通知ID获取所有与之绑定的建筑物信息
// @Tags         Notice-Building
// @Accept       json
// @Produce      json
// @Param        noticeId query int true "通知ID" example:"1"
// @Success      200  {object}  map[string]interface{} "建筑列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的通知ID"
// @Failure      404  {object}  map[string]interface{} "通知不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/notice-building/buildings [get]
// @Security     BearerAuth
func (c *NoticeBuildingController) GetBuildingsByNotice() {
	noticeIDStr := c.Ctx.Query("noticeId")
	if noticeIDStr == "" {
		c.Ctx.JSON(400, gin.H{"error": "noticeId is required"})
		return
	}

	noticeID, err := strconv.ParseUint(noticeIDStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid noticeId"})
		return
	}

	// 检查通知是否存在
	exists, err := c.getService().NoticeExists(uint(noticeID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Notice not found"})
		return
	}

	buildings, err := c.getService().GetBuildingsByNoticeID(uint(noticeID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch buildings"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": buildings})
}

// 4. GetNoticesByBuilding 获取建筑绑定的通知
// @Summary      获取建筑绑定的通知
// @Description  根据建筑ID获取所有与之绑定的通知信息
// @Tags         Notice-Building
// @Accept       json
// @Produce      json
// @Param        buildingId query int true "建筑ID" example:"1"
// @Success      200  {object}  map[string]interface{} "通知列表信息"
// @Failure      400  {object}  map[string]interface{} "无效的建筑ID"
// @Failure      404  {object}  map[string]interface{} "建筑不存在"
// @Failure      500  {object}  map[string]interface{} "服务器内部错误"
// @Router       /admin/notice-building/notices [get]
// @Security     BearerAuth
func (c *NoticeBuildingController) GetNoticesByBuilding() {
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

	// 检查建筑物是否存在
	exists, err := c.getService().BuildingExists(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	if !exists {
		c.Ctx.JSON(404, gin.H{"error": "Building not found"})
		return
	}

	notices, err := c.getService().GetNoticesByBuildingID(uint(buildingID))
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": "Failed to fetch notices"})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": notices})
}
