package base_services

import (
	"fmt"
	"os"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/golang-jwt/jwt/v4"
)

type AuthClaims struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.StandardClaims
}

type IJWTService interface {
	GenerateToken(claims jwt.MapClaims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
	GenerateBuildingAdminToken(admin *base_models.BuildingAdmin) (string, error)
	GenerateDeviceToken(device *base_models.Device) (string, error)
	GenerateSuperAdminToken(admin *base_models.SuperAdmin) (string, error)
}

type JWTService struct {
	secretKey string
}

func NewJWTService() IJWTService {
	// log.Info("初始化JWT服务")
	return &JWTService{
		secretKey: GetSecretKey(),
	}
}

func GetSecretKey() string {
	secret := os.Getenv("SECRET")
	if secret == "" {
		// log.Warn("未设置JWT密钥环境变量，使用默认密钥")
		secret = "secret"
	}
	return secret
}

func (service *JWTService) GenerateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		log.Error("生成令牌失败 | 错误: %v", err)
		return "", err
	}

	log.Debug("令牌生成成功")
	return t, nil
}

func (service *JWTService) ValidateToken(token string) (*jwt.Token, error) {
	log.Debug("验证令牌")
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, isValid := token.Method.(*jwt.SigningMethodHMAC); !isValid {
			err := fmt.Errorf("invalid token %s", token.Header["alg"])
			log.Warn("令牌验证失败，签名方法无效 | 方法: %s", token.Header["alg"])
			return nil, err
		}
		return []byte(service.secretKey), nil
	})
}

func (s *JWTService) GenerateBuildingAdminToken(admin *base_models.BuildingAdmin) (string, error) {
	log.Info("为楼宇管理员生成令牌 | 管理员ID: %d | 邮箱: %s", admin.ID, admin.Email)
	claims := jwt.MapClaims{
		"id":              admin.ID,
		"email":           admin.Email,
		"isBuildingAdmin": true,
		"exp":             time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}

func (s *JWTService) GenerateDeviceToken(device *base_models.Device) (string, error) {
	log.Info("为设备生成令牌 | 设备ID: %s | 建筑ID: %d", device.DeviceID, device.BuildingID)
	claims := jwt.MapClaims{
		"deviceId":   device.DeviceID,
		"buildingId": device.BuildingID,
		"isDevice":   true,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}

func (s *JWTService) GenerateSuperAdminToken(admin *base_models.SuperAdmin) (string, error) {
	log.Info("为超级管理员生成令牌 | 管理员ID: %d | 邮箱: %s", admin.ID, admin.Email)
	claims := jwt.MapClaims{
		"id":      admin.ID,
		"email":   admin.Email,
		"isAdmin": true,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}
