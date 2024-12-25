package databases

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var REDIS_CONN *redis.Client

func InitRedis() error {
	// 从环境变量获取 Redis 配置
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// 转换 DB 编号为整数
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		return fmt.Errorf("invalid redis db number: %v", err)
	}

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	REDIS_CONN = rdb
	return nil
}

func CloseRedis() error {
	if REDIS_CONN != nil {
		return REDIS_CONN.Close()
	}
	return nil
}
