# iBoard 智能楼宇管理系统

## 项目概述

iBoard 是一个基于 Go 语言开发的智能楼宇管理系统，提供了强大的楼宇通知、广告管理和设备控制功能。系统采用 Docker 容器化部署，便于在各种环境中快速安装和更新。

## 系统架构

- **后端**: Go + Gin Web 框架
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **部署**: Docker + Docker Compose
- **通讯**:
  - RESTful API: 基础业务操作
  - 旧系统同步: 与 iSmart 系统保持通知数据同步

## 项目结构

项目采用清晰的分层架构设计：

```
iboard_http_service/
├── cmd/                   # 应用入口点
│   └── server/            # 主服务器
├── internal/              # 内部包，不对外暴露
│   ├── app/               # 应用层
│   │   ├── controller/    # 控制器
│   │   ├── middleware/    # 中间件
│   │   └── router/        # 路由定义
│   ├── domain/            # 领域层
│   │   ├── models/        # 数据模型
│   │   └── services/      # 业务服务
│   │       ├── base/      # 基础服务
│   │       ├── building_admin/ # 楼宇管理员服务
│   │       ├── relationship/ # 关系服务
│   │       └── container/ # 服务容器
│   ├── error/             # 错误处理
│   │   ├── code/          # 错误码
│   │   └── response/      # 响应格式
│   └── infrastructure/    # 基础设施层
│       ├── config/        # 配置管理
│       ├── database/      # 数据库连接池
│       └── redis/         # Redis配置
├── pkg/                   # 可共享的包
├── docs/                  # 文档
│   ├── docs_api/          # API文档
│   ├── docs_deploy/       # 部署文档
│   └── docs_error/        # 错误处理文档
├── scripts/               # 脚本文件
│   ├── deploy/            # 部署脚本
│   ├── migrate/           # 迁移脚本
│   └── update/            # 更新脚本
├── migrations/            # 数据库迁移文件
└── docker-compose.yml     # Docker Compose配置
```

## 主要功能

### 超级管理员功能模块

- **超级管理员账户管理**：添加、登录、更新密码、查询、删除
- **楼宇管理员账户管理**：添加、更新、删除、查询
- **楼宇信息管理**：添加、删除、更新、查询
- **广告管理**：添加、更新、删除、查询
- **通知管理**：添加、批量添加、更新、删除、查询
- **文件管理**：获取上传参数、上传回调、删除、查询
- **楼宇-广告关联管理**：绑定、解绑、查询
- **楼宇-通知关联管理**：绑定、解绑、查询
- **楼宇-管理员关联管理**：绑定、解绑、查询

### 楼宇管理员功能模块

- **广告管理**：添加、更新、删除、查询
- **通知管理**：添加、批量添加、更新、删除、查询
- **文件管理**：获取上传参数、上传回调、删除、查询

### 楼宇功能模块

- **登录**：楼宇登录认证
- **获取广告**：获取分配给楼宇的广告
- **获取通知**：获取分配给楼宇的通知，支持与旧系统同步

## 通知同步流程

系统实现了与旧系统(iSmart)的通知数据同步，主要特点包括：

- **基于MD5的精确比对**：确保内容相同的通知被正确识别，避免重复
- **避免误解绑手动添加的通知**：只解绑旧系统同步的通知
- **并发处理**：动态计算最佳工作线程数，高效处理大量通知
- **缓存优化**：使用Redis缓存建筑物信息、通知ID列表等，提高性能
- **定时同步**：支持定期自动同步和手动触发同步

详细流程请参考 [通知同步流程文档](docs/docs_api/notice_sync_flow.md)

## 部署指南

### 前置要求

- Docker 和 Docker Compose
- Linux 服务器（推荐 Ubuntu 20.04 或 CentOS 8）
- 开放端口：8080(HTTP), 3306(MySQL), 6379(Redis)

### 使用部署脚本

我们提供了部署脚本，可以自动完成部署过程：

1. **配置部署脚本**:

   ```bash
   # 编辑脚本中的服务器信息
   vim scripts/deploy/deploy.sh
   ```

2. **执行部署脚本**:

   ```bash
   chmod +x scripts/deploy/deploy.sh
   ./scripts/deploy/deploy.sh -s 服务器IP -u root
   ```

3. **验证部署**:
   - 访问 `http://服务器IP:8080/api/ping` 检查服务运行状态

### 手动部署

如果需要手动部署，可以按照以下步骤操作：

1. **克隆代码到本地**:

   ```bash
   git clone <repository-url>
   cd iboard_http_service
   ```

2. **配置服务**:

   ```bash
   # 编辑docker-compose.yml文件，根据需要修改配置
   vim docker-compose.yml
   ```

3. **启动服务**:
   ```bash
   docker-compose up -d
   ```

## 迁移指南

当需要将系统从一台服务器迁移到另一台服务器时，可以使用我们提供的迁移脚本：

1. **备份源服务器数据**:

   ```bash
   # 使用迁移脚本
   chmod +x scripts/migrate/simple_migrate.sh
   ./scripts/migrate/simple_migrate.sh
   ```

2. **验证迁移**:
   - 检查服务状态: `docker-compose ps`
   - 测试API接口: `curl http://服务器IP:8080/api/ping`

## 更新指南

当需要更新系统时，可以使用我们提供的更新脚本：

1. **使用更新脚本**:

   ```bash
   # Linux/MacOS
   chmod +x scripts/update/update.sh
   ./scripts/update/update.sh
   
   # Windows
   scripts/update/update.ps1
   ```

2. **直接修复**:

   ```bash
   # 如果需要直接修复某些问题
   chmod +x scripts/update/direct_fix.sh
   ./scripts/update/direct_fix.sh
   ```

## API 文档

主要 API 端点包括：

- **认证**: `/api/auth/login`
- **超级管理员**: `/api/admin/*`
- **楼宇管理员**: `/api/building-admin/*`
- **楼宇**: `/api/buildings/*`
- **广告**: `/api/advertisements/*`
- **通知**: `/api/notices/*`
- **文件**: `/api/files/*`
- **设备**: `/api/devices/*`

## 系统特性

- **自动迁移**: 支持数据库自动迁移
- **基于角色的访问控制**: 不同角色拥有不同权限
- **安全通信**: 基于JWT的API认证
- **性能优化**:
  - 高效的数据库连接池管理
  - Redis缓存机制，提高数据访问速度
  - 通知同步的并发处理
- **容器化部署**: 使用Docker和Docker Compose简化部署和维护
- **旧系统集成**: 与iSmart系统的无缝集成和数据同步

## 故障排除

1. **服务无法启动**:

   - 检查Docker和Docker Compose是否正确安装
   - 检查端口是否被占用: `netstat -tunlp`
   - 查看容器日志: `docker-compose logs app`

2. **数据库连接失败**:

   - 检查数据库配置是否正确
   - 确认数据库服务是否运行: `docker-compose ps db`
   - 检查数据库日志: `docker-compose logs db`

3. **通知同步问题**:

   - 检查Redis缓存配置
   - 确认与旧系统的连接正常
   - 查看同步日志

## 规范设置

### 1. 代码风格规范

- **Go 语言规范**: 遵循官方 Go 语言规范和最佳实践
- **命名规则**:
  - 包名: 小写单词，不使用下划线或混合大小写
  - 文件名: 小写，使用下划线分隔多个单词
  - 结构体名: 驼峰命名法，首字母大写
  - 接口名: 通常以 "er" 结尾，如 Reader, Writer
  - 方法和函数: 驼峰命名法，公开方法首字母大写，私有方法首字母小写

### 2. 接口设计规范

- **RESTful API 设计**:
  - 使用恰当的 HTTP 方法: GET(查询), POST(创建), PUT(更新), DELETE(删除)
  - 路径使用名词而非动词
  - 使用复数形式表示资源集合
- **响应格式统一**:
  ```json
  {
    "code": 200,
    "message": "操作成功",
    "data": { ... }
  }
  ```

### 3. 错误管理规范

- **错误码系统**: 使用统一的错误码体系，便于问题定位和客户端处理
- **错误处理流程**:
  - 在发生错误的地方捕获并记录详细信息
  - 向上层返回有意义的错误信息
  - 在 API 层统一格式化错误响应

## 许可证

版权所有 © 2024 iBoard 开发团队 

# iBoard HTTP Service

## Docker镜像版本管理

本项目支持为Docker镜像添加版本号标签，方便追踪和管理不同版本的部署。

### 构建带版本号的Docker镜像

使用提供的构建脚本可以生成带有指定版本号的Docker镜像：

```bash
# 使用默认版本号(1.0.0)构建
./scripts/build/docker-build.sh

# 指定版本号构建
./scripts/build/docker-build.sh 1.2.3
```

构建完成后，将生成格式为`iboard_http_service:版本号`的Docker镜像。

### 在docker-compose中使用版本号

也可以直接通过环境变量设置版本号：

```bash
# 设置版本号环境变量
export VERSION=1.2.3

# 构建并启动服务
docker-compose up -d --build
```

### 查看镜像版本号

可以通过以下命令查看镜像的版本号标签：

```bash
docker inspect iboard_http_service:1.0.0 | grep -A 3 Labels
``` 