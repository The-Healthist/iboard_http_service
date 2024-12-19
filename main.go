package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/router"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

func initSuperAdmin(db *gorm.DB) error {
	var admin base_models.SuperAdmin

	// 检查是否已存在管理员
	result := db.Where("email = ?", "admin@example.com").First(&admin)
	if result.Error == nil {
		// 已存在管理员，不需要创建
		return nil
	}

	// 创建默认密码的哈希值
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("admin123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// 创建默认管理员
	newAdmin := base_models.SuperAdmin{
		Email:    "admin@example.com",
		Password: string(hashedPassword),
	}

	if err := db.Create(&newAdmin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %v", err)
	}

	log.Println("Default super admin created successfully")
	return nil
}

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Init Email
	emailPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		fmt.Println("smtp_port error")
		return
	}
	utils.InitEmail(os.Getenv("SMTP_ADDR"), int(emailPort), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))

	// Init DB
	db := databases.InitDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	if db == nil {
		log.Fatal("Failed to initialize database")
	}

	// 初始化超级管理员
	if err := initSuperAdmin(db); err != nil {
		log.Printf("Failed to initialize super admin: %v\n", err)
	}

	r := gin.Default()
	router.RegisterRoute(r)

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
