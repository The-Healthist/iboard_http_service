package databases

import (
	"fmt"

	"github.com/The-Healthist/iboard_http_service/internal/domain/models"
	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB_CONN *gorm.DB

func InitDB(host, user, password, port, dbname string) *gorm.DB {
	log.Info("开始初始化数据库连接...")

	// 先连接MySQL（不指定数据库）
	rootDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	log.Debug("连接MySQL服务器: %s:%s", host, port)

	db, err := gorm.Open(mysql.Open(rootDsn), &gorm.Config{})
	if err != nil {
		log.Error("连接MySQL服务器失败: %v", err)
		return nil
	}
	log.Debug("MySQL服务器连接成功")

	// 创建数据库
	createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbname)
	log.Debug("尝试创建数据库: %s", dbname)

	if err := db.Exec(createDB).Error; err != nil {
		log.Error("创建数据库失败: %v", err)
		return nil
	}
	log.Debug("数据库创建或已存在: %s", dbname)

	// 连接指定的数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)
	log.Debug("连接到指定数据库: %s", dbname)

	DB_CONN, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("连接数据库失败: %v", err)
		return nil
	}
	log.Info("数据库连接成功: %s", dbname)

	// 添加自动迁移
	log.Info("开始数据库表结构迁移...")
	if err := DB_CONN.AutoMigrate(
		&models.SuperAdmin{},
		&models.BuildingAdmin{},
		&models.Building{},
		&models.Advertisement{},
		&models.Notice{},
		&models.File{},
		&models.Device{}, // 包含 JSON 列：top/full/notices carousel lists
		&models.App{},
	); err != nil {
		log.Error("数据库表结构迁移失败: %v", err)
		return nil
	}

	log.Info("数据库表结构迁移完成")
	return DB_CONN
}

func AutoMigrate(db *gorm.DB) error {
	log.Info("开始数据库表结构迁移...")

	// 在这里添加所有需要迁移的模型
	err := db.AutoMigrate(
		&models.SuperAdmin{},
		&models.BuildingAdmin{},
		&models.Building{},
		&models.Advertisement{},
		&models.Notice{},
		&models.File{},
		&models.Device{},
		&models.App{},
	)

	if err != nil {
		log.Error("数据库表结构迁移失败: %v", err)
		return fmt.Errorf("auto migration failed: %v", err)
	}

	log.Info("数据库表结构迁移完成")
	return nil
}
