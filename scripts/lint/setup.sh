#!/bin/bash

# 获取项目根目录
ROOT_DIR=$(git rev-parse --show-toplevel)

# 安装golangci-lint
echo "检查golangci-lint是否已安装..."
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint未安装，正在安装..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    if [ $? -ne 0 ]; then
        echo "安装golangci-lint失败，请手动安装：https://golangci-lint.run/usage/install/"
        exit 1
    fi
    
    echo "golangci-lint安装成功"
else
    echo "golangci-lint已安装"
fi

# 添加执行权限
chmod +x $ROOT_DIR/scripts/lint/run_lint.sh

echo "静态代码检查工具设置完成！"
echo "你可以通过运行 ./scripts/lint/run_lint.sh 来执行代码检查" 