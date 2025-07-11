#!/bin/bash

# 新服务器设置
NEW_HOST="117.72.193.54"
NEW_PORT="22"
NEW_USER="root"
NEW_PASS="1090119your@"
NEW_DIR="/root/iboard"

echo "开始修复数据库连接问题..."

# 连接到新服务器并执行修复操作
export SSHPASS="$NEW_PASS"
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && \
echo '1. 备份配置文件...' && \
cp .env .env.bak && cp docker-compose.yml docker-compose.yml.bak && \
echo '2. 修改.env文件中的数据库主机名...' && \
sed -i 's/DB_HOST=.*/DB_HOST=mysql/g' .env && \
echo '3. 检查环境变量中的数据库配置...' && \
grep -A 5 DB_ .env && \
echo '4. 修改docker-compose.yml文件中的后端服务配置...' && \
cat > fix_backend.sh << 'EOF'
#!/bin/bash
# 修改后端服务的环境变量配置
sed -i '/backend:/,/networks:/ s/DB_HOST=.*/DB_HOST=mysql/' docker-compose.yml
sed -i '/backend:/,/networks:/ s/DB_PORT=.*/DB_PORT=3306/' docker-compose.yml

# 确保后端服务依赖于mysql和redis
if grep -q 'depends_on:' docker-compose.yml; then
  # 已有depends_on，检查是否包含mysql和redis
  if ! grep -q 'depends_on:.*mysql' docker-compose.yml; then
    sed -i '/depends_on:/a\\      - mysql' docker-compose.yml
  fi
  if ! grep -q 'depends_on:.*redis' docker-compose.yml; then
    sed -i '/depends_on:/a\\      - redis' docker-compose.yml
  fi
else
  # 没有depends_on，添加
  sed -i '/backend:/,/networks:/ s/networks:/depends_on:\\n      - mysql\\n      - redis\\n    networks:/' docker-compose.yml
fi

echo '后端服务配置修复完成'
EOF
chmod +x fix_backend.sh && ./fix_backend.sh && \
echo '5. 停止并重新启动服务...' && \
docker-compose down && \
docker-compose up -d && \
echo '6. 等待服务启动...' && \
sleep 15 && \
echo '7. 检查服务状态...' && \
docker-compose ps && \
echo '8. 检查后端日志...' && \
docker-compose logs --tail=20 backend"

echo "修复操作完成。"
echo "如果问题仍然存在，请尝试以下操作："
echo "1. 检查MySQL容器是否正常启动"
echo "2. 检查网络配置是否正确"
echo "3. 检查数据库用户名和密码是否正确"
echo "4. 尝试手动连接数据库：docker-compose exec mysql mysql -u root -p" 