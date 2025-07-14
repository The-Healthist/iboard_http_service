# Swagger API 文档使用指南

## 简介

本项目使用 Swagger 自动生成 API 文档。Swagger 是一种用于描述 RESTful API 的规范，能够提供交互式的 API 文档，方便开发人员理解和测试 API。

## Swagger 注解规范

在 Go 代码中添加 Swagger 注释时，遵循以下规范：

### 控制器方法注释格式

```go
// [序号]. [方法名] [中文描述]
// @Summary      [方法简短描述]
// @Description  [方法详细描述]
// @Tags         [API 分组标签]
// @Accept       [接受的数据格式，如 json]
// @Produce      [返回的数据格式，如 json]
// @Param        [参数名] [参数位置] [参数类型] [是否必需] [参数描述] example:"[示例值]"
// @Success      [成功状态码]  {[返回数据类型]}  [返回数据模型] "[返回说明]"
// @Failure      [失败状态码]  {[返回数据类型]}  [返回数据模型] "[错误说明]"
// @Router       [API路径] [HTTP方法]
// @Security     [认证方式]//是否需要token,是token还是device—token
```

### 参数位置

- `path`: URL 路径参数，如 `/users/{id}`
- `query`: URL 查询参数，如 `/users?page=1`
- `body`: 请求体参数
- `header`: HTTP 头参数
- `formData`: 表单数据

### 示例

```go
// 1. GetUser 获取用户信息
// @Summary      获取用户信息
// @Description  根据用户ID获取用户详细信息
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        id path int true "用户ID" example:"1"
// @Success      200  {object}  map[string]interface{} "用户信息"
// @Failure      404  {object}  map[string]interface{} "用户不存在"
// @Router       /api/users/{id} [get]
// @Security     BearerAuth
```

## 如何添加 Swagger 注释

1. 为每个控制器方法添加符合上述规范的注释
2. 添加的注释应包含以下内容：
   - 方法编号和名称
   - 简要和详细描述
   - API 分组标签
   - 请求和响应格式
   - 参数描述（包括示例值）
   - 成功和失败响应
   - API 路由和 HTTP 方法
   - 认证方式

## 生成 Swagger 文档

执行以下命令生成 Swagger 文档：

```bash
swag init -g cmd/server/main.go -o docs/docs_swagger
```

## 访问 Swagger UI

项目运行后，可通过以下 URL 访问 Swagger UI：

```
http://localhost:10031/swagger/index.html
```

在 Swagger UI 中，你可以：

1. 浏览所有 API 端点
2. 查看 API 详细文档
3. 测试 API 请求
4. 查看请求和响应示例

## 认证

对于需要认证的 API，在 Swagger UI 中：

1. 点击页面右上角的 "Authorize" 按钮
2. 输入 Bearer Token（格式为: `Bearer your_token_here`）
3. 点击 "Authorize" 进行授权

## 最佳实践

1. 为每个控制器方法添加序号和中文描述，方便团队成员快速理解
2. 提供完整的参数描述和示例值
3. 对于复杂参数，使用示例说明其格式
4. 描述所有可能的响应状态码和相应的数据结构
5. 保持文档与代码同步更新
