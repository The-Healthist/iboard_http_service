# 日志系统使用文档

本文档详细介绍了iBoard HTTP服务的日志系统，包括如何在代码中使用日志、如何配置日志以及Docker环境中的日志处理。

## 目录

1. [日志包概述](#日志包概述)
2. [基本用法](#基本用法)
3. [日志级别](#日志级别)
4. [配置选项](#配置选项)
5. [与Gin框架集成](#与Gin框架集成)
6. [Docker环境中的日志配置](#Docker环境中的日志配置)
7. [最佳实践](#最佳实践)
8. [路径处理](#路径处理)
9. [Gin框架集成增强](#Gin框架集成增强)

## 日志包概述

iBoard HTTP服务使用自定义的日志包(`pkg/log`)，它提供了以下功能：

- 多级别日志记录（DEBUG、INFO、WARN、ERROR、FATAL）
- 日志文件自动轮转
- 日志文件自动清理
- 多输出目标支持（控制台、文件等）
- 与Gin框架的集成

## 基本用法

### 导入日志包

```go
import "github.com/The-Healthist/iboard_http_service/pkg/log"
```

### 使用全局日志函数

```go
// 记录不同级别的日志
log.Debug("这是一条调试日志，参数: %v", someVar)
log.Info("这是一条信息日志")
log.Warn("这是一条警告日志")
log.Error("这是一条错误日志")
log.Fatal("这是一条致命错误日志，程序将退出") // 会导致程序退出
```

### 创建自定义日志记录器

```go
// 创建自定义日志记录器
logger, err := log.NewLogger(
    log.WithLevel(log.DEBUG),
    log.WithLogDir("custom_logs"),
    log.WithMaxFileSize(2*1024*1024), // 2MB
)
if err != nil {
    panic(err)
}
defer logger.Close()

// 使用自定义日志记录器
logger.Info("使用自定义日志记录器")
```

## 日志级别

日志系统支持以下日志级别（从低到高）：

1. **DEBUG**: 调试信息，用于开发和排查问题
2. **INFO**: 一般信息，表示正常的系统操作
3. **WARN**: 警告信息，表示可能的问题或异常情况
4. **ERROR**: 错误信息，表示发生了错误但程序可以继续运行
5. **FATAL**: 致命错误，记录日志后程序将退出

只有等于或高于设置级别的日志才会被记录。例如，如果设置级别为INFO，则DEBUG级别的日志不会被记录。

### 设置日志级别

```go
// 设置全局日志记录器的级别
log.SetLevel(log.DEBUG)

// 或者在创建自定义日志记录器时设置
logger, _ := log.NewLogger(log.WithLevel(log.DEBUG))
```

## 配置选项

创建日志记录器时可以使用以下选项：

```go
logger, err := log.NewLogger(
    // 设置日志级别
    log.WithLevel(log.INFO),
    
    // 设置日志文件目录
    log.WithLogDir("logs"),
    
    // 设置单个日志文件最大大小（字节）
    log.WithMaxFileSize(1 * 1024 * 1024), // 1MB
    
    // 设置最大保留的日志文件数量
    log.WithMaxFiles(365),
    
    // 设置日志文件时间格式
    log.WithTimeFormat("2006-01-02"),
    
    // 添加额外的输出目标
    log.WithWriter(customWriter),
)
```

## 与Gin框架集成

日志包提供了与Gin框架集成的中间件：

```go
import (
    "github.com/The-Healthist/iboard_http_service/pkg/log"
    "github.com/gin-gonic/gin"
)

func main() {
    // 初始化日志
    if err := log.InitLogger(log.WithLevel(log.INFO)); err != nil {
        panic(err)
    }
    
    // 设置Gin的默认日志输出
    log.SetGinDefaultLogger()
    
    // 创建Gin引擎
    r := gin.New()
    
    // 使用日志中间件
    r.Use(log.GinLoggerMiddleware())
    
    // 使用恢复中间件
    r.Use(log.GinRecoveryMiddleware())
    
    // 注册路由...
    
    // 启动服务器
    r.Run(":10031")
}
```

## Docker环境中的日志配置

在Docker环境中，日志文件被挂载到宿主机，便于查看和管理。

### docker-compose.yml中的日志配置

```yaml
services:
  backend:
    build: .
    container_name: iboard_http_service
    restart: always
    ports:
      - '10031:10031'
    environment:
      - TZ=Asia/Shanghai
      # 其他环境变量...
    volumes:
      - ./logs:/app/logs  # 将容器内的日志目录挂载到宿主机
    depends_on:
      - mysql
      - redis
    networks:
      - iboard-network
```

### 查看Docker容器日志

除了应用程序生成的日志文件外，还可以使用Docker命令查看容器的标准输出日志：

```bash
# 查看容器日志
docker logs iboard_http_service

# 实时查看日志
docker logs -f iboard_http_service

# 查看最近100行日志
docker logs --tail 100 iboard_http_service
```

## 日志路径

日志文件默认保存在项目根目录下的`logs`文件夹中，文件名格式为`app_YYYY-MM-DD.log`。

> **注意**：之前的版本可能会在`cmd/server/logs`目录下也生成日志文件。现在已经统一到项目根目录的`logs`文件夹中。如果您发现两个位置都有日志文件，可以运行`./scripts/cleanup/cleanup_logs.sh`脚本来清理旧的日志目录。

## 最佳实践

### 1. 合理选择日志级别

- **生产环境**: 建议使用INFO级别，只记录重要信息
- **开发环境**: 可以使用DEBUG级别，记录更多调试信息
- **测试环境**: 可以根据需要选择INFO或DEBUG级别

### 2. 结构化日志内容

记录日志时，提供足够的上下文信息：

```go
// 不好的例子
log.Error("操作失败")

// 好的例子
log.Error("用户登录失败: 用户ID=%d, 原因=%s", userID, reason)
```

### 3. 避免过多日志

避免在高频执行的代码中记录过多日志，特别是DEBUG级别的日志，这可能会影响性能。

### 4. 定期检查和清理日志

虽然日志系统会自动清理旧日志文件，但仍建议定期检查日志目录，确保磁盘空间充足。

### 5. 在关键操作点记录日志

在以下关键点记录日志：

- 应用程序启动和关闭
- 用户登录和注销
- 重要的业务操作
- 系统错误和异常情况
- 性能瓶颈

### 6. 保护敏感信息

确保日志中不包含敏感信息，如密码、令牌等。如果需要记录包含敏感信息的对象，先进行脱敏处理。 

## 路径处理

日志系统现在支持智能路径处理：

1. 当使用相对路径（如`logs`）时，日志文件会保存在项目根目录下的相应文件夹中
2. 当使用绝对路径时，日志文件会保存在指定的绝对路径中
3. 系统会自动检测当前工作目录，确保无论从哪里启动程序，日志都保存在正确的位置

## Gin框架集成增强

日志系统与Gin框架的集成已增强，提供以下功能：

1. **请求ID跟踪**：每个HTTP请求都会生成唯一的请求ID，并在响应头中返回（`X-Request-ID`）
2. **用户ID记录**：如果在请求上下文中设置了用户ID，日志会自动记录
3. **错误分离**：请求错误会单独记录，方便排查问题
4. **Panic恢复**：自动捕获并记录panic，确保服务稳定性

### 在控制器中设置用户ID示例

```go
func YourHandler(c *gin.Context) {
    // 从JWT或会话中获取用户ID
    userID := getUserIDFromToken(c)
    
    // 设置到上下文中，日志中间件会自动记录
    c.Set(log.UserIDKey, userID)
    
    // 处理请求...
}
```