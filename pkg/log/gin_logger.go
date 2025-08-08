package log

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// 请求ID的键名
	RequestIDKey = "RequestID"
	// 用户ID的键名
	UserIDKey = "UserID"
)

// GinLoggerMiddleware 返回一个Gin中间件，用于记录HTTP请求日志
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求ID
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 请求延迟
		latency := endTime.Sub(startTime)

		// 请求方法
		method := c.Request.Method

		// 请求路径
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		// 状态码
		statusCode := c.Writer.Status()

		// 客户端IP
		clientIP := c.ClientIP()

		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 获取用户ID（如果存在）
		userID, exists := c.Get(UserIDKey)
		userIDStr := "-"
		if exists {
			userIDStr = toString(userID)
		}

		// 根据状态码选择日志级别
		var logFunc func(string, ...interface{})
		switch {
		case statusCode >= 500:
			logFunc = Error
		case statusCode >= 400:
			logFunc = Warn
		default:
			logFunc = Info
		}

		// 记录日志
		logFunc("[GIN] %s | %s | %s | %d | %s | %s | %s | %s",
			requestID,
			startTime.Format("2006/01/02 - 15:04:05"),
			clientIP,
			statusCode,
			latency,
			method,
			path,
			userIDStr,
		)

		// 如果有错误，单独记录
		if errorMessage != "" {
			Error("[GIN] 请求错误: %s | %s", requestID, errorMessage)
		}
	}
}

// GinRecoveryMiddleware 返回一个Gin中间件，用于捕获panic并记录错误日志
func GinRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求ID
				requestID, exists := c.Get(RequestIDKey)
				requestIDStr := "-"
				if exists {
					requestIDStr = requestID.(string)
				}

				// 记录请求信息
				path := c.Request.URL.Path
				method := c.Request.Method

				// 记录错误日志
				Error("[GIN] 请求处理发生panic: %s | %s | %s | 错误: %v",
					requestIDStr,
					method,
					path,
					err)

				// 返回500响应
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

// SetGinDefaultLogger 设置Gin框架的默认日志输出
func SetGinDefaultLogger() {
	// 将Gin的默认日志输出设置为我们的日志记录器
	gin.DefaultWriter = defaultLogger.logFile
	if defaultLogger.logFile != nil {
		gin.DefaultWriter = defaultLogger.logFile
	}
}

// 将任意类型转为字符串
func toString(v interface{}) string {
	if v == nil {
		return "-"
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case uint:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case uint64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
