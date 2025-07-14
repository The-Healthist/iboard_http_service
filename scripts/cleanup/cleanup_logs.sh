#!/bin/bash

# 获取项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/../.." && pwd)

# 定义日志目录
MAIN_LOG_DIR="$PROJECT_ROOT/logs"
OLD_LOG_DIR="$PROJECT_ROOT/cmd/server/logs"

echo "开始清理日志目录..."

# 创建主日志目录（如果不存在）
if [ ! -d "$MAIN_LOG_DIR" ]; then
  echo "创建主日志目录: $MAIN_LOG_DIR"
  mkdir -p "$MAIN_LOG_DIR"
fi

# 检查旧日志目录是否存在
if [ -d "$OLD_LOG_DIR" ]; then
  echo "检测到旧日志目录: $OLD_LOG_DIR"
  
  # 检查旧日志目录中是否有日志文件
  if [ "$(ls -A "$OLD_LOG_DIR")" ]; then
    echo "旧日志目录中存在文件，将移动到主日志目录"
    
    # 移动旧日志文件到主日志目录
    mv "$OLD_LOG_DIR"/* "$MAIN_LOG_DIR"/ 2>/dev/null
    echo "旧日志文件已移动到: $MAIN_LOG_DIR"
  fi
  
  # 删除旧的空日志目录
  rm -rf "$OLD_LOG_DIR"
  echo "旧日志目录已删除"
fi

echo "日志清理完成。所有日志将统一保存在: $MAIN_LOG_DIR" 