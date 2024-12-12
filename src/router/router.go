package router

import (
	"github.com/gin-gonic/gin"
	"github.com/your-project/controllers"
	"github.com/your-project/middlewares"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(middlewares.Cors())

	// API 路由组
	api := router.Group("/api")
	{
		// 无需认证的路由
		api.POST("/login", controllers.Login)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middlewares.AuthMiddleware())
		{
			// 建筑相关
			building := authorized.Group("/buildings")
			{
				building.GET("", controllers.GetBuildings)
				building.POST("", controllers.CreateBuilding)
				building.PUT("/:id", controllers.UpdateBuilding)
				building.DELETE("/:id", controllers.DeleteBuilding)
			}

			// 其他路由...
		}
	}

	return router
}
