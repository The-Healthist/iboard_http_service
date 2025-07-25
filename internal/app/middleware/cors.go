package middlewares

import (
	"net/http"

	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		log.Debug("处理跨域请求 | %v | 方法: %s | 来源: %s", requestID, method, origin)

		// 无论是否有Origin头，都设置CORS头
		// 接收客户端发送的origin （重要！）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		//服务器支持的所有跨域请求的方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		//允许跨域设置可以返回其他子段，可以自定义字段
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Authorization")
		// 允许浏览器（客户端）可以解析的头部 （重要）
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		//设置缓存时间
		c.Header("Access-Control-Max-Age", "172800")
		//允许客户端传递校验信息比如 cookie (重要)
		c.Header("Access-Control-Allow-Credentials", "true")

		//允许类型校验
		if method == "OPTIONS" {
			log.Debug("预检请求处理 | %v | 来源: %s", requestID, origin)
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Error("中间件发生Panic | %v | 错误: %v", requestID, err)
			}
		}()

		c.Next()
	}
}
