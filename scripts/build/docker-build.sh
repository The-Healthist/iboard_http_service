#!/bin/bash

# 默认版本号
DEFAULT_VERSION="1.0.0"

# 获取版本号参数
VERSION=${1:-$DEFAULT_VERSION}

echo "=== 开始构建 iBoard HTTP Service 镜像 ==="
echo "版本号: $VERSION"

# 设置环境变量
export VERSION=$VERSION

# 构建Docker镜像
docker-compose build

echo "=== Docker镜像构建完成 ==="
echo "镜像名称: iboard_http_service:$VERSION"

# 显示构建的镜像
echo "=== 镜像信息 ==="
docker images | grep iboard_http_service 