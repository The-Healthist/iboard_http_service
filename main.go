package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/router"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

// setupLogging configures the logging for both file and console output
func setupLogging() (*os.File, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create log file with timestamp
	currentTime := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("logs/app_%s.log", currentTime)
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Set up multi-writer for both console and file output
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(gin.DefaultWriter)

	return logFile, nil
}

func initSuperAdmin(db *gorm.DB) error {
	var admin base_models.SuperAdmin

	log.Println("Checking for existing super admin...")
	result := db.Where("email = ?", "admin@example.com").First(&admin)
	if result.Error == nil {
		log.Println("Super admin already exists")
		return nil
	}

	log.Println("Creating default super admin...")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("admin123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

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
	// Set up logging
	logFile, err := setupLogging()
	if err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		return
	}
	defer logFile.Close()

	log.Println("Starting iBoard HTTP Server...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Environment variables loaded successfully")

	// Init Email
	log.Println("Initializing email service...")
	emailPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		log.Fatal("SMTP port configuration error:", err)
	}
	utils.InitEmail(os.Getenv("SMTP_ADDR"), int(emailPort), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	log.Println("Email service initialized successfully")

	// Init DB
	log.Println("Initializing database connection...")
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
	log.Println("Database connection established successfully")

	// Initialize super admin
	if err := initSuperAdmin(db); err != nil {
		log.Printf("Failed to initialize super admin: %v\n", err)
	}

	// Initialize Redis connection
	log.Println("Initializing Redis connection...")
	if err := databases.InitRedis(); err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
	defer databases.CloseRedis()
	log.Println("Redis connection established successfully")

	// Configure Gin
	log.Println("Configuring Gin framework...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add custom logging middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format
		return fmt.Sprintf("[GIN] %s | %s | %d | %s | %s | %s | %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())

	log.Println("Registering routes...")
	router.RegisterRoute(r)
	log.Println("Routes registered successfully")

	// Start server
	serverAddr := "0.0.0.0:10031"
	log.Printf("Starting HTTP server on %s...\n", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
