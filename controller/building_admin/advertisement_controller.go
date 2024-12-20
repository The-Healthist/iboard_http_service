package building_admin_controllers

import (
	"strconv"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/utils/response"
	"github.com/gin-gonic/gin"
)

type BuildingAdminAdvertisementController struct {
	ctx     *gin.Context
	service building_admin_services.InterfaceBuildingAdminAdvertisementService
}

func NewBuildingAdminAdvertisementController(
	ctx *gin.Context,
	service building_admin_services.InterfaceBuildingAdminAdvertisementService,
) *BuildingAdminAdvertisementController {
	return &BuildingAdminAdvertisementController{
		ctx:     ctx,
		service: service,
	}
}

func (c *BuildingAdminAdvertisementController) GetAdvertisements() {
	email := c.ctx.GetString("email")

	query := make(map[string]interface{})
	if adType := c.ctx.Query("type"); adType != "" {
		query["type"] = adType
	}

	paginate := map[string]interface{}{
		"pageSize": 10,
		"pageNum":  1,
		"desc":     true,
	}

	advertisements, pagination, err := c.service.Get(email, query, paginate)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"data":       advertisements,
		"pagination": pagination,
	})
}

func (c *BuildingAdminAdvertisementController) GetAdvertisement() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisement ID"})
		return
	}

	advertisement, err := c.service.GetByID(uint(id), email)
	if err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"data": advertisement})
}

type CreateAdvertisementDTO struct {
	Title       string    `json:"title" binding:"required,min=2,max=100"`
	Content     string    `json:"content" binding:"required"`
	StartTime   time.Time `json:"startTime" binding:"required"`
	EndTime     time.Time `json:"endTime" binding:"required,gtfield=StartTime"`
	FileID      uint      `json:"fileId" binding:"required"`
	BuildingIDs []uint    `json:"buildingIds" binding:"required,min=1"`
}

func (c *BuildingAdminAdvertisementController) CreateAdvertisement() {
	var dto CreateAdvertisementDTO
	if err := c.ctx.ShouldBindJSON(&dto); err != nil {
		response.ValidationError(c.ctx, err)
		return
	}

	email := c.ctx.GetString("email")

	var advertisement base_models.Advertisement
	if err := c.ctx.ShouldBindJSON(&advertisement); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Create(&advertisement, email); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{
		"message": "Advertisement created successfully",
		"data":    advertisement,
	})
}

func (c *BuildingAdminAdvertisementController) UpdateAdvertisement() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisement ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ctx.ShouldBindJSON(&updates); err != nil {
		c.ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.Update(uint(id), email, updates); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Advertisement updated successfully"})
}

func (c *BuildingAdminAdvertisementController) DeleteAdvertisement() {
	email := c.ctx.GetString("email")
	idStr := c.ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.ctx.JSON(400, gin.H{"error": "Invalid advertisement ID"})
		return
	}

	if err := c.service.Delete(uint(id), email); err != nil {
		c.ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.ctx.JSON(200, gin.H{"message": "Advertisement deleted successfully"})
}
