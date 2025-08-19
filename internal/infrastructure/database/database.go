package databases

import (
	"fmt"
	"time"

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

	// 第一步：先创建versions表（不依赖其他表）
	if err := DB_CONN.AutoMigrate(&models.Version{}); err != nil {
		log.Error("创建versions表失败: %v", err)
		return nil
	}
	log.Info("versions表创建成功")

	// 第二步：初始化默认版本数据
	if err := initDefaultVersionData(DB_CONN); err != nil {
		log.Error("初始化默认版本数据失败: %v", err)
		return nil
	}

	// 确保 apps 表存在
	if err := DB_CONN.AutoMigrate(&models.App{}); err != nil {
		log.Error("创建apps表失败: %v", err)
		return nil
	}
	log.Info("apps表创建或已存在")

	// 第三步：处理apps表，先添加字段，再建立外键关系
	if err := migrateAppsTable(DB_CONN); err != nil {
		log.Error("迁移apps表失败: %v", err)
		return nil
	}

	// 第四步：迁移其他表，包括 App 模型以确保 apps 表存在
	if err := DB_CONN.AutoMigrate(
		&models.App{},
		&models.SuperAdmin{},
		&models.BuildingAdmin{},
		&models.Building{},
		&models.Advertisement{},
		&models.Notice{},
		&models.File{},
		&models.Device{}, // 包含 JSON 列：top/full/notices carousel lists
	); err != nil {
		log.Error("迁移其他表失败: %v", err)
		return nil
	}

	log.Info("数据库表结构迁移完成")
	return DB_CONN
}

func AutoMigrate(db *gorm.DB) error {
	log.Info("开始数据库表结构迁移...")

	// 第一步：先创建versions表（不依赖其他表）
	if err := db.AutoMigrate(&models.Version{}); err != nil {
		log.Error("创建versions表失败: %v", err)
		return fmt.Errorf("create versions table failed: %v", err)
	}
	log.Info("versions表创建成功")

	// 第二步：迁移versions表字段
	if err := migrateVersionsTable(db); err != nil {
		log.Error("迁移versions表字段失败: %v", err)
		return fmt.Errorf("migrate versions table failed: %v", err)
	}
	log.Info("versions表字段迁移完成")

	// 第二步：初始化默认版本数据
	if err := initDefaultVersionData(db); err != nil {
		log.Error("初始化默认版本数据失败: %v", err)
		return fmt.Errorf("init default version data failed: %v", err)
	}

	// 第三步：处理apps表，先添加字段，再建立外键关系
	if err := migrateAppsTable(db); err != nil {
		log.Error("迁移apps表失败: %v", err)
		return fmt.Errorf("migrate apps table failed: %v", err)
	}

	// 第四步：迁移其他表
	err := db.AutoMigrate(
		&models.SuperAdmin{},
		&models.BuildingAdmin{},
		&models.Building{},
		&models.Advertisement{},
		&models.Notice{},
		&models.File{},
		&models.Device{},
	)

	if err != nil {
		log.Error("迁移其他表失败: %v", err)
		return fmt.Errorf("migrate other tables failed: %v", err)
	}

	log.Info("数据库表结构迁移完成")
	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData(db *gorm.DB) error {
	log.Info("开始初始化默认数据...")

	// 检查是否已有版本数据
	var count int64
	if err := db.Model(&models.Version{}).Count(&count).Error; err != nil {
		log.Error("检查版本数据失败: %v", err)
		return err
	}

	// 如果没有版本数据，创建默认版本
	if count == 0 {
		log.Info("创建默认版本数据...")
		defaultVersion := &models.Version{
			VersionNumber: "1.0.0",
			BuildNumber:   "001",
			Description:   "Initial version",
			DownloadUrl:   "",
			Status:        "active",
		}

		if err := db.Create(defaultVersion).Error; err != nil {
			log.Error("创建默认版本失败: %v", err)
			return err
		}
		log.Info("默认版本创建成功: %s", defaultVersion.VersionNumber)
	}

	// 检查是否已有App配置数据
	var appCount int64
	if err := db.Model(&models.App{}).Count(&appCount).Error; err != nil {
		log.Error("检查App配置数据失败: %v", err)
		return err
	}

	// 如果没有App配置数据，创建默认配置
	if appCount == 0 {
		log.Info("创建默认App配置...")
		defaultApp := &models.App{
			CurrentVersionID: 1, // 关联到第一个版本
			LastCheckTime:    time.Now(),
			UpdateInterval:   3600,
			AutoUpdate:       false,
			Status:           "active",
		}

		if err := db.Create(defaultApp).Error; err != nil {
			log.Error("创建默认App配置失败: %v", err)
			return err
		}
		log.Info("默认App配置创建成功")
	}

	log.Info("默认数据初始化完成")
	return nil
}

// initDefaultVersionData 初始化默认版本数据
func initDefaultVersionData(db *gorm.DB) error {
	log.Info("开始初始化默认版本数据...")

	// 检查是否已有版本数据
	var count int64
	if err := db.Model(&models.Version{}).Count(&count).Error; err != nil {
		log.Error("检查版本数据失败: %v", err)
		return err
	}

	// 如果没有版本数据，创建默认版本
	if count == 0 {
		log.Info("创建默认版本数据...")
		defaultVersion := &models.Version{
			VersionNumber: "1.0.0",
			BuildNumber:   "001",
			Description:   "Initial version",
			DownloadUrl:   "",
			Status:        "active",
		}

		if err := db.Create(defaultVersion).Error; err != nil {
			log.Error("创建默认版本失败: %v", err)
			return err
		}
		log.Info("默认版本创建成功: %s", defaultVersion.VersionNumber)
	}

	log.Info("默认版本数据初始化完成")
	return nil
}

// migrateAppsTable 迁移apps表，处理外键约束
func migrateAppsTable(db *gorm.DB) error {
	log.Info("开始迁移apps表...")

	// 检查apps表是否存在version字段（旧字段）
	var hasOldVersionField bool
	err := db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'apps' AND COLUMN_NAME = 'version'").Scan(&hasOldVersionField).Error
	if err != nil {
		log.Error("检查旧version字段失败: %v", err)
		return err
	}

	// 如果存在旧version字段，先删除它
	if hasOldVersionField {
		log.Info("发现旧version字段，正在删除...")
		if err := db.Exec("ALTER TABLE apps DROP COLUMN version").Error; err != nil {
			log.Error("删除旧version字段失败: %v", err)
			return err
		}
		log.Info("旧version字段删除成功")
	}

	// 检查apps表是否存在current_version_id字段
	var hasColumn bool
	err = db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'apps' AND COLUMN_NAME = 'current_version_id'").Scan(&hasColumn).Error
	if err != nil {
		log.Error("检查字段失败: %v", err)
		return err
	}

	if !hasColumn {
		log.Info("添加新字段到apps表...")

		// 添加新字段（不包含外键约束）
		if err := db.Exec("ALTER TABLE apps ADD COLUMN current_version_id bigint unsigned DEFAULT 1 COMMENT '当前使用的版本ID'").Error; err != nil {
			log.Error("添加current_version_id字段失败: %v", err)
			return err
		}

		if err := db.Exec("ALTER TABLE apps ADD COLUMN last_check_time datetime(3) DEFAULT NULL COMMENT '最后检查更新时间'").Error; err != nil {
			log.Error("添加last_check_time字段失败: %v", err)
			return err
		}

		if err := db.Exec("ALTER TABLE apps ADD COLUMN update_interval int DEFAULT 3600 COMMENT '检查更新间隔(秒)'").Error; err != nil {
			log.Error("添加update_interval字段失败: %v", err)
			return err
		}

		if err := db.Exec("ALTER TABLE apps ADD COLUMN auto_update boolean DEFAULT false COMMENT '是否自动更新'").Error; err != nil {
			log.Error("添加auto_update字段失败: %v", err)
			return err
		}

		if err := db.Exec("ALTER TABLE apps ADD COLUMN status varchar(50) DEFAULT 'active' COMMENT '应用状态'").Error; err != nil {
			log.Error("添加status字段失败: %v", err)
			return err
		}

		log.Info("新字段添加成功")
	}

	// 检查是否已有App配置数据
	var appCount int64
	if err := db.Model(&models.App{}).Count(&appCount).Error; err != nil {
		log.Error("检查App配置数据失败: %v", err)
		return err
	}

	// 如果没有App配置数据，创建默认配置
	if appCount == 0 {
		log.Info("创建默认App配置...")
		defaultApp := &models.App{
			CurrentVersionID: 1, // 关联到第一个版本
			LastCheckTime:    time.Now(),
			UpdateInterval:   3600,
			AutoUpdate:       false,
			Status:           "active",
		}

		if err := db.Create(defaultApp).Error; err != nil {
			log.Error("创建默认App配置失败: %v", err)
			return err
		}
		log.Info("默认App配置创建成功")
	}

	// 现在可以安全地添加外键约束
	log.Info("添加外键约束...")
	if err := db.Exec("ALTER TABLE apps ADD CONSTRAINT fk_apps_current_version_id FOREIGN KEY (current_version_id) REFERENCES versions(id) ON DELETE SET NULL").Error; err != nil {
		// 如果外键已存在，忽略错误
		log.Info("外键约束可能已存在或添加失败，继续执行...")
	}

	// 添加索引
	if err := db.Exec("ALTER TABLE apps ADD INDEX idx_apps_current_version_id (current_version_id)").Error; err != nil {
		// 如果索引已存在，忽略错误
		log.Info("索引可能已存在或添加失败，继续执行...")
	}

	log.Info("apps表迁移完成")
	return nil
}

// migrateVersionsTable 迁移versions表，处理字段变更
func migrateVersionsTable(db *gorm.DB) error {
	log.Info("开始迁移versions表...")

	// 检查并删除旧字段
	fieldsToDrop := []string{
		"release_date",
		"is_force_update",
		"file_size",
		"min_version",
	}

	for _, field := range fieldsToDrop {
		var hasField bool
		query := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'versions' AND COLUMN_NAME = '%s'", field)

		if err := db.Raw(query).Scan(&hasField).Error; err != nil {
			log.Error("检查字段 %s 失败: %v", field, err)
			continue
		}

		if hasField {
			log.Info("删除旧字段: %s", field)
			dropQuery := fmt.Sprintf("ALTER TABLE versions DROP COLUMN %s", field)
			if err := db.Exec(dropQuery).Error; err != nil {
				log.Error("删除字段 %s 失败: %v", field, err)
				continue
			}
			log.Info("字段 %s 删除成功", field)
		}
	}

	// 检查build_number字段是否存在
	var hasBuildNumber bool
	query := "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'versions' AND COLUMN_NAME = 'build_number'"

	if err := db.Raw(query).Scan(&hasBuildNumber).Error; err != nil {
		log.Error("检查build_number字段失败: %v", err)
		return err
	}

	// 如果build_number字段不存在，添加它
	if !hasBuildNumber {
		log.Info("添加build_number字段...")
		if err := db.Exec("ALTER TABLE versions ADD COLUMN build_number varchar(50) DEFAULT '001' COMMENT '构建号'").Error; err != nil {
			log.Error("添加build_number字段失败: %v", err)
			return err
		}
		log.Info("build_number字段添加成功")
	}

	log.Info("versions表迁移完成")
	return nil
}
