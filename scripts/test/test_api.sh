#!/bin/bash

# 测试API登录接口
echo "测试超级管理员登录接口..."
curl -X 'POST' \
  'http://localhost:10031/api/admin/login' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "email": "admin@example.com",
  "password": "admin123"
}'

echo -e "\n\n测试完成" 