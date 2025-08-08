# 日志系统迁移指南

本文档提供了将代码从标准库日志迁移到新日志包的指南。

## 迁移步骤概述

1. 导入新的日志包
2. 替换标准库日志调用
3. 利用新日志包的高级功能
4. 测试日志输出

## 详细步骤

### 1. 导入新的日志包

在每个需要使用日志的文件中，添加以下导入语句：

```go
import "github.com/The-Healthist/iboard_http_service/pkg/log"
```

如果文件中已经导入了标准库的`log`包，请移除它：

```go
// 移除这一行
import "log"
```

### 2. 替换标准库日志调用

使用以下对照表替换日志调用：

| 标准库调用 | 新日志包调用 | 说明 |
|----------|------------|------|
| `log.Printf("信息: %v", val)` | `log.Info("信息: %v", val)` | 信息级别日志 |
| `log.Println("错误")` | `log.Error("错误")` | 错误级别日志 |
| `fmt.Printf("警告: %v\n", err)` | `log.Warn("警告: %v", err)` | 警告级别日志 |
| `log.Fatalf("致命错误: %v", err)` | `log.Fatal("致命错误: %v", err)` | 致命错误日志 |

注意：新日志包中不需要在格式字符串末尾添加`\n`，日志系统会自动添加换行符。

### 3. 利用新日志包的高级功能

#### 3.1 在中间件中使用请求ID和用户ID

在中间件中，可以获取请求ID并在日志中使用：

```go
// 获取请求ID
requestID, _ := c.Get(log.RequestIDKey)

// 在日志中使用
log.Info("处理请求 | %v | 路径: %s", requestID, path)
```

设置用户ID以便在日志中跟踪：

```go
// 从JWT或会话中获取用户ID
userID := getUserIDFromToken(c)

// 设置到上下文中，日志中间件会自动记录
c.Set(log.UserIDKey, userID)
```

#### 3.2 使用不同的日志级别

根据信息的重要性使用不同的日志级别：

```go
// 调试信息
log.Debug("这是调试信息，值: %v", someValue)

// 一般信息
log.Info("操作成功完成")

// 警告信息
log.Warn("发现潜在问题: %v", warning)

// 错误信息
log.Error("操作失败: %v", err)

// 致命错误（会导致程序退出）
log.Fatal("无法继续执行: %v", fatalErr)
```

### 4. 测试日志输出

修改完成后，运行应用并检查日志输出是否符合预期：

- 日志文件应该保存在项目根目录的`logs`文件夹中
- 日志应该包含适当的级别标记（如`[INFO]`、`[ERROR]`等）
- 日志格式应该一致

## 迁移示例

### 示例1：数据库连接

```go
// 旧代码
if err := db.AutoMigrate(&models.User{}); err != nil {
    log.Printf("Failed to auto migrate: %v\n", err)
    return nil
}
log.Println("Database migration completed successfully")

// 新代码
if err := db.AutoMigrate(&models.User{}); err != nil {
    log.Error("数据库自动迁移失败: %v", err)
    return nil
}
log.Info("数据库迁移完成")
```

### 示例2：中间件

```go
// 旧代码
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            log.Println("No token provided")
            c.AbortWithStatus(401)
            return
        }
        // ...
    }
}

// 新代码
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取请求ID
        requestID, _ := c.Get(log.RequestIDKey)
        
        token := c.GetHeader("Authorization")
        if token == "" {
            log.Warn("未提供认证令牌 | %v", requestID)
            c.AbortWithStatus(401)
            return
        }
        
        // ...
    }
}
```

## 已完成的迁移

以下文件已经完成迁移：

1. `internal/app/middleware/jwt.go`
2. `internal/app/middleware/cors.go`
3. `internal/infrastructure/database/database.go`
4. `internal/infrastructure/redis/redis.go`
5. `internal/domain/services/container/service_container.go`
6. `internal/app/controller/base/upload_control.go`
7. `internal/domain/services/base/notice_sync_service.go`
8. `internal/domain/services/base/building_service.go`
9. `internal/app/controller/base/device_controller.go`
10. `internal/domain/services/building_admin/file_building_admin_service.go`
11. `internal/domain/services/relationship/device_building_service.go`
12. `internal/domain/services/relationship/notice_building_service.go`
13. `internal/domain/services/relationship/file_advertisement_service.go`
14. `internal/domain/services/relationship/file_notice_service.go`

## 待迁移的文件

以下是需要继续迁移的文件：

1. 控制器文件：
   - 其他 `internal/app/controller/base/` 下的控制器
   - `internal/app/controller/building_admin/`
   - `internal/app/controller/relationship/`

2. 服务文件：
   - 其他 `internal/domain/services/base/` 下的服务
   - 其他 `internal/domain/services/building_admin/` 下的服务
   - 其他 `internal/domain/services/relationship/` 下的服务

## 注意事项

1. 确保在每个文件中只导入一次日志包
2. 注意格式字符串中的占位符使用
3. 根据信息的重要性选择适当的日志级别
4. 在中间件中尽可能使用请求ID和用户ID
5. 保持日志消息的一致性，优先使用中文消息 