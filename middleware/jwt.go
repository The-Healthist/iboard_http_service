package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)

			if claims["isAdmin"] == false {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			c.Set("email", claims["email"])
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}

func AuthorizeJWTStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)

			if claims["isAdmin"] == true {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			c.Set("email", claims["email"])
			c.Next()
		} else {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}
