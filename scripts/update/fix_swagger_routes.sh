#!/bin/bash

# 脚本用于修复Swagger路由路径，移除@Router注解中的/api前缀

echo "开始修复Swagger路由路径..."

# 查找所有控制器文件
CONTROLLERS_DIR="internal/app/controller"
FILES=$(find $CONTROLLERS_DIR -type f -name "*.go")

# 计数器
FIXED_COUNT=0
PROCESSED_COUNT=0

# 处理每个文件
for FILE in $FILES; do
  PROCESSED_COUNT=$((PROCESSED_COUNT + 1))
  echo "处理文件: $FILE"
  
  # 使用sed替换@Router /api/ 为 @Router /
  # 创建临时文件
  sed 's|// @Router[ \t]*/api/|// @Router       /|g' "$FILE" > "$FILE.tmp"
  
  # 检查是否有修改
  if diff "$FILE" "$FILE.tmp" > /dev/null; then
    echo "  - 无需修改"
    rm "$FILE.tmp"
  else
    FIXED_COUNT=$((FIXED_COUNT + 1))
    echo "  - 修复了路由路径"
    mv "$FILE.tmp" "$FILE"
  fi
done

echo "完成！处理了 $PROCESSED_COUNT 个文件，修复了 $FIXED_COUNT 个文件的路由路径。"
echo "请运行 'swag init -g cmd/server/main.go -o docs/docs_swagger' 重新生成Swagger文档。" 