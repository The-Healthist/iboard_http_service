package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/your-project/services"
	"github.com/your-project/utils"
)

type BuildingController struct {
	service *services.BuildingService
}

func NewBuildingController() *BuildingController {
	return &BuildingController{
		service: services.NewBuildingService(),
	}
}

func (c *BuildingController) GetBuildings(ctx *gin.Context) {
	// 参数验证
	var query struct {
		Page     int    `form:"page" binding:"required,min=1"`
		PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
		Name     string `form:"name"`
	}

	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	// 调用service
	buildings, total, err := c.service.GetBuildings(query.Page, query.PageSize, query.Name)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseSuccess(ctx, gin.H{
		"items": buildings,
		"total": total,
	})
}
