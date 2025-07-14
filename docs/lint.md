# 静态代码检查工具使用指南

本项目使用 [golangci-lint](https://golangci-lint.run/) 作为静态代码检查工具，以确保代码质量和一致性。

## 安装与配置

1. 运行安装脚本：

```bash
./scripts/lint/setup.sh
```

这个脚本会：
- 安装 golangci-lint（如果尚未安装）
- 配置必要的权限

## 手动运行代码检查

可以随时手动运行代码检查：

```bash
./scripts/lint/run_lint.sh
```

## 配置文件说明

配置文件位于项目根目录的 `.golangci.yml`，主要包含以下配置：

### 运行配置

```yaml
run:
  timeout: 5m  # 超时设置
  skip-dirs:   # 忽略的目录
    - logs
    - scripts
    - docs
    - migrations
  skip-files:  # 忽略的文件
    - _test.go
    - .env
    # ...其他忽略的文件类型
```

### 启用的检查器

```yaml
linters:
  disable-all: true
  enable:
    - errcheck
    - gosimple
    - govet
    # ...其他启用的检查器
```

## 常见问题及修复方法

1. **godot**: 注释应该以句点结束
   ```go
   // 这是一个注释 // 错误
   // 这是一个注释。 // 正确
   ```

2. **gofmt/goimports**: 代码格式问题
   ```bash
   # 自动修复
   gofmt -w ./path/to/file.go
   goimports -w ./path/to/file.go
   ```

3. **未使用的变量/导入**
   - 删除未使用的变量和导入，或者在必要时使用 `_` 标记

4. **函数过长**
   - 将长函数拆分为多个小函数，提高代码可读性和可维护性

## 如何忽略特定检查

在特殊情况下，可以在代码中添加注释来忽略特定的检查：

```go
// nolint:errcheck
someFunc() // 忽略错误检查

// nolint
func someFunc() {} // 忽略所有检查

// nolint:errcheck,gosimple
someFunc() // 忽略多个检查
```

**注意**: 请谨慎使用这些忽略标记，只在确实必要的情况下使用。 