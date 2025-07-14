package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/The-Healthist/iboard_http_service/pkg/log"
	"github.com/go-redis/redis/v8"
)

var REDIS_CONN *redis.Client

func InitRedis() error {
	var err error
	maxRetries := 5
	retryDelay := time.Second * 3

	log.Info("开始初始化Redis连接...")

	for i := 0; i < maxRetries; i++ {
		err = initRedisConnection()
		if err == nil {
			log.Info("Redis连接成功 (尝试 %d/%d)", i+1, maxRetries)
			return nil
		}

		log.Error("Redis连接失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Info("将在 %v 后重试...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("经过 %d 次尝试后，Redis连接仍然失败: %v", maxRetries, err)
}

func initRedisConnection() error {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis" // 默认使用容器名
		log.Debug("未设置REDIS_HOST环境变量，使用默认值: %s", redisHost)
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
		log.Debug("未设置REDIS_PORT环境变量，使用默认值: %s", redisPort)
	}

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0"
		log.Debug("未设置REDIS_DB环境变量，使用默认值: %s", redisDB)
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
		log.Debug("未设置REDIS_PASSWORD环境变量，使用空密码")
	}

	db, err := strconv.Atoi(redisDB)
	if err != nil {
		log.Error("Redis数据库编号无效: %v", err)
		return fmt.Errorf("无效的Redis数据库编号: %v", err)
	}

	log.Info("尝试连接Redis: %s:%s (DB: %d)", redisHost, redisPort, db)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:     redisPassword,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("Redis连接测试失败: %v", err)
		return fmt.Errorf("Redis连接测试失败: %v", err)
	}

	log.Debug("Redis连接测试成功")
	REDIS_CONN = rdb
	return nil
}

func CloseRedis() error {
	if REDIS_CONN != nil {
		log.Info("关闭Redis连接...")
		err := REDIS_CONN.Close()
		if err != nil {
			log.Error("关闭Redis连接时出错: %v", err)
		} else {
			log.Debug("Redis连接已成功关闭")
		}
		return err
	}
	return nil
}
