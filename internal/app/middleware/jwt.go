package middlewares

import (
	"net/http"
	"strings"

	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func extractToken(authHeader string) string {
	// 检查并移除 "Bearer " 前缀
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}
	return authHeader
}

func AuthorizeJWTAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Warn("认证失败: 无效的令牌声明 | %v", requestID)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "invalid token claims",
				})
				return
			}

			if claims["isAdmin"] == false {
				log.Warn("认证失败: 非管理员尝试访问 | %v | %v", requestID, claims["email"])
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			// 设置用户ID用于日志记录
			if email, ok := claims["email"].(string); ok {
				c.Set(log.UserIDKey, email)
			}

			// Set claims directly without conversion
			c.Set("email", claims["email"])
			c.Set("claims", claims)

			log.Debug("管理员认证成功 | %v | %v", requestID, claims["email"])
			c.Next()
		} else {
			log.Warn("认证失败: 无效令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}

func AuthorizeJWTStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)

			if claims["isAdmin"] == true {
				log.Warn("认证失败: 管理员尝试访问员工资源 | %v | %v", requestID, claims["email"])
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			// 设置用户ID用于日志记录
			if email, ok := claims["email"].(string); ok {
				c.Set(log.UserIDKey, email)
			}

			c.Set("email", claims["email"])

			log.Debug("员工认证成功 | %v | %v", requestID, claims["email"])
			c.Next()
		} else {
			log.Error("认证失败: 无效令牌 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}

func AuthorizeJWTBuildingAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			log.Warn("认证失败: 无效的令牌格式 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn("认证失败: 无效的令牌声明 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isBuildingAdmin"] != true {
			log.Warn("认证失败: 非建筑管理员尝试访问 | %v | %v", requestID, claims["email"])
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// 设置用户ID用于日志记录
		if email, ok := claims["email"].(string); ok {
			c.Set(log.UserIDKey, email)
		}

		c.Set("email", claims["email"])

		log.Debug("建筑管理员认证成功 | %v | %v", requestID, claims["email"])
		c.Next()
	}
}

func AuthorizeJWTBuilding() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			log.Warn("认证失败: 无效的令牌格式 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn("认证失败: 无效的令牌声明 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isBuilding"] != true {
			log.Warn("认证失败: 非建筑用户尝试访问 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// Convert jwt.MapClaims to map[string]interface{} before setting in context
		claimsMap := make(map[string]interface{})
		for key, value := range claims {
			claimsMap[key] = value
		}

		// 设置用户ID用于日志记录
		if buildingID, ok := claims["buildingID"].(float64); ok {
			c.Set(log.UserIDKey, int(buildingID))
		}

		// Set claims in context for later use
		c.Set("claims", claimsMap)

		log.Debug("建筑用户认证成功 | %v | BuildingID: %v", requestID, claims["buildingID"])
		c.Next()
	}
}

// AuthorizeJWTUpload is a middleware that authorizes both superadmin and buildingadmin tokens
func AuthorizeJWTUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			log.Warn("认证失败: 无效的令牌格式 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn("认证失败: 无效的令牌声明 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Check if the token is either for a superadmin or buildingadmin
		isAdmin, _ := claims["isAdmin"].(bool)
		isBuildingAdmin, _ := claims["isBuildingAdmin"].(bool)

		if !isAdmin && !isBuildingAdmin {
			log.Warn("认证失败: 权限不足 | %v | %v", requestID, claims["email"])
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// 设置用户ID用于日志记录
		if email, ok := claims["email"].(string); ok {
			c.Set(log.UserIDKey, email)
		}

		// Set claims in context for later use
		c.Set("claims", claims)
		c.Set("email", claims["email"])

		userType := "超级管理员"
		if isBuildingAdmin {
			userType = "建筑管理员"
		}

		log.Debug("上传认证成功 | %v | %s: %v", requestID, userType, claims["email"])
		c.Next()
	}
}

func AuthorizeJWTDevice() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求ID
		requestID, _ := c.Get(log.RequestIDKey)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("认证失败: 未提供令牌 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			log.Warn("认证失败: 无效的令牌格式 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			log.Warn("认证失败: 令牌验证错误 | %v | %v", requestID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Warn("认证失败: 无效的令牌声明 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isDevice"] != true {
			log.Warn("认证失败: 非设备用户尝试访问 | %v", requestID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// Convert jwt.MapClaims to map[string]interface{} before setting in context
		claimsMap := make(map[string]interface{})
		for key, value := range claims {
			claimsMap[key] = value
		}

		// 设置用户ID用于日志记录
		if deviceID, ok := claims["deviceID"].(string); ok {
			c.Set(log.UserIDKey, deviceID)
		}

		// Set claims in context for later use
		c.Set("claims", claimsMap)

		log.Debug("设备认证成功 | %v | DeviceID: %v", requestID, claims["deviceID"])
		c.Next()
	}
}
