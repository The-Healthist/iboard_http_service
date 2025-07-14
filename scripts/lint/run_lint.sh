#!/bin/bash

# 检查golangci-lint是否已安装
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint未安装，正在安装..."
    # 安装golangci-lint
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    if [ $? -ne 0 ]; then
        echo "安装golangci-lint失败，请手动安装：https://golangci-lint.run/usage/install/"
        exit 1
    fi
    
    echo "golangci-lint安装成功"
fi

echo "开始运行代码静态检查..."

# 运行golangci-lint
golangci-lint run ./...

# 检查运行结果
if [ $? -eq 0 ]; then
    echo "代码检查通过，没有发现问题"
else
    echo "代码检查发现问题，请修复上述问题"
fi 