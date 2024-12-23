# iBoard 后端接口设计文档
## 架构设计
### 分层架构
系统采用经典的四层架构设计：
- Controller 层：处理 HTTP 请求响应
- Service 层：实现业务逻辑
- Repository 层：数据访问层
- Model 层：数据模型定义
### 每层职责
#### Controller 层
- 处理 HTTP 请求和响应
- 参数校验和绑定
- 调用 Service 层处理业务逻辑
- 统一响应格式处理
- 错误处理和状态码管理
#### Service 层
- 实现核心业务逻辑
- 调用 Repository 层进行数据操作
- 数据组装和转换
- 事务管理
- 业务规则验证
#### Repository 层
- 数据库操作接口
- 实现数据的 CRUD 操作
- 数据持久化
- 查询优化
- 数据缓存处理
#### Model 层
- 定义数据结构
- 定义表关系
- 字段验证规则
- 模型关联关系
## 模块划分
系统分为基础模块(base)和关系模块(relationship)两大类：

## SuperAdmin 功能模块
### 1. Super Admin Management
- 超级管理员账户管理
  - 登录认证
  - 密码修改
  - 个人信息管理
### 2. Building Admin Management
- 楼宇管理员账户管理
  - 创建管理员账户
  - 修改管理员信息
  - 删除管理员账户
  - 管理员状态控制
  - 管理员列表查询
### 3. Building Management
- 楼宇信息管理
  - 创建楼宇信息
  - 修改楼宇信息
  - 删除楼宇信息
  - 楼宇状态控制
  - 楼宇列表查询
### 4. BuildingAdmin-Building Relationship Management
- 管理员与楼宇关联管理
  - 绑定管理员到楼宇
  - 解绑管理员与楼宇
  - 查询楼宇的管理员列表
  - 查询管理员的楼宇列表
### 5. File Management
- 文件管理
  - 文件上传
  - 文件下载
  - 文件删除
  - 文件信息修改
  - 文件列表查询
### 6. Advertisement Management
- 广告管理
  - 创建广告
  - 修改广告信息
  - 删除广告
  - 广告状态控制
  - 广告列表查询
### 7. Notice Management
- 通知管理
  - 创建通知
  - 修改通知信息
  - 删除通知
  - 通知状态控制
  - 通知列表查询
### 8. API Management
- API接口管理
  - API密钥管理
  - 接口权限控制
  - 接口调用监控
### 9. Building Data Async
- 楼宇数据同步
  - 数据同步配置
  - 同步状态监控
  - 同步日志查看
## BuildingAdmin 功能模块
## BuildingAdmin login 功能为下面的接口添加token
## 登录以后添加token才能操作下面的接口
### 1. File Management
  - 文件上传
  - 文件查看
  - 文件下载
  - 文件列表查询
  - 文件删除(仅仅能管理自己上传的文件)
  - 文件修改(仅仅能管理自己上传的文件)
### 2. Advertisement Management
- 广告管理
  - 查看广告列表
  - 广告删除(仅仅能管理自己上传的文件)
  - 广告修改(仅仅能管理自己上传的文件)
  - 广告添加
### 3. Notice Management
- 通知管理
  - 查看通知列表
  - 添加通知
  - 删除通知(仅仅能管理自己上传的文件)
  - 修改通知(仅仅能管理自己上传的文件)