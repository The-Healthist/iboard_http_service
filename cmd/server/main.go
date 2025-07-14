package main

// @title           ILock HTTP Service API
// @version         1.0
// @description     智能门禁管理系统API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.yourcompany.com/support
// @contact.email  support@yourcompany.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:10031
// @BasePath  /api

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 输入带有Bearer前缀的用户认证token

// @securityDefinitions.apikey  DeviceAuth
// @in                          header
// @name                        Authorization
// @description                 输入带有Bearer前缀的设备认证token

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/The-Healthist/iboard_http_service/docs/docs_swagger" // Import Swagger docs
	"github.com/The-Healthist/iboard_http_service/internal/app/router"
	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	base_services "github.com/The-Healthist/iboard_http_service/internal/domain/services/base"
	"github.com/The-Healthist/iboard_http_service/internal/domain/services/container"
	databases "github.com/The-Healthist/iboard_http_service/internal/infrastructure/database"
	"github.com/The-Healthist/iboard_http_service/internal/infrastructure/redis"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/The-Healthist/iboard_http_service/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func initSuperAdmin(db *gorm.DB) error {
	var admin models.SuperAdmin

	log.Info("检查超级管理员是否存在...")
	result := db.Where("email = ?", "admin@example.com").First(&admin)
	if result.Error == nil {
		log.Info("超级管理员已存在")
		return nil
	}

	log.Info("创建默认超级管理员...")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("admin123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %v", err)
	}

	newAdmin := models.SuperAdmin{
		Email:    "admin@example.com",
		Password: string(hashedPassword),
	}

	if err := db.Create(&newAdmin).Error; err != nil {
		return fmt.Errorf("创建管理员失败: %v", err)
	}

	log.Info("默认超级管理员创建成功")
	return nil
}

func main() {
	// 初始化日志系统（简化版）
	if err := log.InitLogger(log.WithLogDir("logs")); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		return
	}
	defer log.GetLogger().Close()

	log.Info("启动iBoard HTTP服务器...")

	// 尝试加载.env文件 - 首先尝试当前目录
	if err := godotenv.Load(); err != nil {
		// 如果当前目录没有找到，尝试从项目根目录加载
		rootEnvPath := "../.env"
		if err := godotenv.Load(rootEnvPath); err != nil {
			// 尝试从绝对路径加载
			workDir, _ := os.Getwd()
			absRootPath := workDir + "/../../.env"
			if err := godotenv.Load(absRootPath); err != nil {
				log.Error("加载.env文件失败，尝试了当前目录、项目根目录和绝对路径")
				return
			}
		}
	}
	log.Info("环境变量加载成功")

	// 初始化Email服务
	log.Info("初始化邮件服务...")
	emailPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		log.Fatal("SMTP端口配置错误: %v", err)
	}
	utils.InitEmail(os.Getenv("SMTP_ADDR"), int(emailPort), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	log.Info("邮件服务初始化成功")

	// 初始化数据库
	log.Info("初始化数据库连接...")
	db := databases.InitDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	if db == nil {
		log.Fatal("初始化数据库失败")
	}
	log.Info("数据库连接建立成功")

	// 初始化超级管理员
	if err := initSuperAdmin(db); err != nil {
		log.Error("初始化超级管理员失败: %v", err)
	}

	// 初始化Redis连接
	log.Info("初始化Redis连接...")
	maxRedisRetries := 5
	var redisErr error
	for i := 0; i < maxRedisRetries; i++ {
		redisErr = redis.InitRedis()
		if redisErr == nil {
			log.Info("Redis连接建立成功")
			break
		}
		log.Warn("初始化Redis失败(尝试 %d/%d): %v", i+1, maxRedisRetries, redisErr)
		if i < maxRedisRetries-1 {
			time.Sleep(time.Second * 3)
		}
	}
	if redisErr != nil {
		log.Fatal("多次尝试后仍无法初始化Redis: %v", redisErr)
	}

	// 确保Redis连接在程序退出时关闭
	defer func() {
		if err := redis.CloseRedis(); err != nil {
			log.Error("关闭Redis连接时出错: %v", err)
		}
	}()

	// 初始化服务容器
	log.Info("初始化服务容器...")
	serviceContainer := container.NewServiceContainer(db)
	log.Info("服务容器初始化成功")

	// 配置Gin
	log.Info("配置Gin框架...")
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎，使用自定义日志中间件
	r := gin.New()
	r.Use(log.GinLoggerMiddleware())
	r.Use(log.GinRecoveryMiddleware())

	log.Info("注册路由...")
	router.RegisterRoute(r)
	log.Info("路由注册成功")

	// 启动通知同步调度器
	log.Info("启动通知同步调度器...")
	ctx := context.Background()
	noticeSyncService := serviceContainer.GetService("noticeSync").(base_services.InterfaceNoticeSyncService)
	noticeSyncService.StartSyncScheduler(ctx)
	log.Info("通知同步调度器启动成功")

	// 启动服务器
	serverAddr := "0.0.0.0:10031"
	log.Info("启动HTTP服务器，监听地址: %s...", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("启动服务器失败: %v", err)
	}
}
