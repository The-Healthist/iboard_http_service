#!/bin/bash

# 设置工作目录为项目根目录
cd "$(dirname "$0")/../.." || exit

# 控制器目录
CONTROLLER_DIRS=(
  "internal/app/controller/base"
  "internal/app/controller/building_admin"
  "internal/app/controller/relationship"
)

echo "开始更新Swagger路由注解，添加/api前缀..."

# 遍历所有控制器目录
for dir in "${CONTROLLER_DIRS[@]}"; do
  echo "处理目录: $dir"
  
  # 查找所有控制器文件
  for file in "$dir"/*.go; do
    echo "  处理文件: $file"
    
    # 使用sed替换@Router注解，添加/api前缀
    # 注意：这里假设所有@Router路径都以/admin, /building_admin, /device开头
    sed -i'.bak' -E 's|// @Router[[:space:]]+/admin/|// @Router       /api/admin/|g' "$file"
    sed -i'.bak' -E 's|// @Router[[:space:]]+/building_admin/|// @Router       /api/building_admin/|g' "$file"
    sed -i'.bak' -E 's|// @Router[[:space:]]+/device/|// @Router       /api/device/|g' "$file"
  done
done

# 删除备份文件
find "${CONTROLLER_DIRS[@]}" -name "*.bak" -delete

echo "路由注解更新完成！"
echo "请执行 'swag init -g cmd/server/main.go -o docs/docs_swagger' 重新生成Swagger文档" 