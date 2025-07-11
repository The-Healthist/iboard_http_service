#!/bin/bash

# 服务器配置
ORIGINAL_HOST="39.108.49.167"  # 原始服务器
ORIGINAL_PORT="22"
ORIGINAL_USER="root"
ORIGINAL_PASS="1090119your@"
ORIGINAL_DIR="/root/iboard"

NEW_HOST="117.72.193.54"  # 新服务器
NEW_PORT="22"
NEW_USER="root"
NEW_PASS="1090119your@"
NEW_DIR="/root/iboard"

# 备份目录
BACKUP_DIR="./backup"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/iboard_backup_${TIMESTAMP}.tar.gz"

# 颜色输出函数
function print_info() {
  echo -e "\033[0;34m[INFO] $1\033[0m"
}

function print_success() {
  echo -e "\033[0;32m[SUCCESS] $1\033[0m"
}

function print_warning() {
  echo -e "\033[0;33m[WARNING] $1\033[0m"
}

function print_error() {
  echo -e "\033[0;31m[ERROR] $1\033[0m"
}

# 错误处理函数
function handle_error() {
  print_error "$1"
  print_info "清理临时文件..."
  rm -rf temp_migrate
  rm -f iboard_migrate.tar.gz
  exit 1
}

# 检查sshpass是否安装
if ! command -v sshpass &> /dev/null; then
  print_warning "sshpass未安装，将尝试安装..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install sshpass || { 
      print_error "sshpass安装失败！请手动安装: brew install sshpass"; 
      exit 1; 
    }
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt-get update && sudo apt-get install -y sshpass || { print_error "sshpass安装失败！请手动安装: sudo apt-get install sshpass"; exit 1; }
  else
    print_error "无法识别的操作系统，请手动安装sshpass后重试"; 
    exit 1;
  fi
  print_success "sshpass安装成功"
fi

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 检查服务器连接
print_info "检查原始服务器连接..."
export SSHPASS="$ORIGINAL_PASS"
if ! sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "echo '连接到原始服务器成功'"; then
  handle_error "无法连接到原始服务器，请检查连接信息"
fi

print_info "检查新服务器连接..."
export SSHPASS="$NEW_PASS"
if ! sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "echo '连接到新服务器成功'"; then
  handle_error "无法连接到新服务器，请检查连接信息"
fi

# 步骤1: 检查原始服务器上的服务状态
print_info "步骤1: 检查原始服务器上的服务状态..."
export SSHPASS="$ORIGINAL_PASS"
sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && docker-compose ps" || handle_error "无法检查服务状态"

# 步骤2: 保存原始服务器上的Docker镜像
print_info "步骤2: 保存原始服务器上的Docker镜像..."
export SSHPASS="$ORIGINAL_PASS"

# 获取后端镜像名称
BACKEND_IMAGE=$(sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && docker-compose ps -q backend | xargs docker inspect -f '{{.Config.Image}}' 2>/dev/null || echo 'iboard_backend'")
print_info "后端镜像名称: $BACKEND_IMAGE"

# 保存镜像为tar文件
print_info "保存镜像为tar文件..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && docker save -o backend_image.tar $BACKEND_IMAGE" || print_warning "无法保存Docker镜像，将尝试使用源代码构建"

# 步骤3: 从原始服务器复制所有必要的文件
print_info "步骤3: 从原始服务器复制所有必要的文件..."

# 创建临时目录
mkdir -p temp_migrate
cd temp_migrate || handle_error "无法创建临时目录"

# 从原始服务器下载关键文件
print_info "从原始服务器下载关键文件..."
export SSHPASS="$ORIGINAL_PASS"

# 下载docker-compose.yml
print_info "下载docker-compose.yml..."
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/docker-compose.yml" ./ || handle_error "无法下载docker-compose.yml"

# 下载Dockerfile
print_info "下载Dockerfile..."
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/Dockerfile" ./ || print_warning "Dockerfile不存在，将创建默认Dockerfile"

# 如果Dockerfile不存在，创建一个默认的
if [ ! -f "Dockerfile" ]; then
  print_info "创建默认Dockerfile..."
  cat > Dockerfile << 'EOF'
FROM golang:1.20 as build-stage

# Set working directory
WORKDIR /app

# Copy source code
COPY . /app

# Build the application
RUN go build -o main .

# Run release
FROM golang:1.20

WORKDIR /app

COPY --from=build-stage /app/main /app/
COPY --from=build-stage /app/migrations /app/migrations

EXPOSE 10031

CMD ["/app/main"]
EOF
fi

# 下载.env文件
print_info "下载.env文件..."
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/.env" ./ || print_warning ".env文件不存在，请确保手动上传"

# 备份数据库
print_info "备份MySQL数据库..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && docker-compose exec -T mysql mysqldump -u root -p1090119your --all-databases > iboard_db_backup.sql" || handle_error "无法备份MySQL数据库"
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/iboard_db_backup.sql" ./ || handle_error "无法下载数据库备份"

# 备份Redis数据
print_info "备份Redis数据..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && docker-compose exec -T redis redis-cli SAVE && docker cp iboard_redis:/data/dump.rdb ./redis_dump.rdb" || print_warning "无法备份Redis数据"
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/redis_dump.rdb" ./ || print_warning "无法下载Redis备份"

# 下载Docker镜像（如果存在）
if sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && [ -f backend_image.tar ]"; then
  print_info "下载Docker镜像..."
  sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/backend_image.tar" ./ || print_warning "无法下载Docker镜像"
fi

# 下载源代码
print_info "下载源代码..."
# 创建源代码目录
mkdir -p controller middleware models router services utils database migrations

# 下载关键目录
for dir in controller middleware models router services utils database migrations; do
  print_info "下载 $dir 目录..."
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && tar -czf ${dir}.tar.gz $dir" || print_warning "无法打包 $dir 目录"
  sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/${dir}.tar.gz" ./ || print_warning "无法下载 $dir 目录"
  
  if [ -f "${dir}.tar.gz" ]; then
    tar -xzf ${dir}.tar.gz || print_warning "无法解压 $dir 目录"
    rm ${dir}.tar.gz
    sshpass -e ssh -o StrictHostKeyChecking=no -p "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST" "cd $ORIGINAL_DIR && rm ${dir}.tar.gz" || print_warning "无法删除远程临时文件"
  fi
done

# 下载go.mod和go.sum
print_info "下载go.mod和go.sum..."
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/go.mod" ./ || print_warning "go.mod不存在，请确保手动上传"
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/go.sum" ./ || print_warning "go.sum不存在，请确保手动上传"

# 下载main.go
print_info "下载main.go..."
sshpass -e scp -o StrictHostKeyChecking=no -P "$ORIGINAL_PORT" "$ORIGINAL_USER@$ORIGINAL_HOST:$ORIGINAL_DIR/main.go" ./ || print_warning "main.go不存在，请确保手动上传"

# 步骤4: 创建本地备份
print_info "步骤4: 创建本地备份..."
cd ..
print_info "创建备份文件: $BACKUP_FILE"
tar -czf "$BACKUP_FILE" -C temp_migrate . || handle_error "无法创建备份文件"
print_success "备份文件已创建: $BACKUP_FILE"

# 步骤5: 更新.env文件中的回调URL
cd temp_migrate || handle_error "无法进入临时目录"
if [ -f ".env" ]; then
  print_info "步骤5: 更新.env文件中的回调URL..."
  sed -i.bak "s|CALLBACK_URL=http://[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}|CALLBACK_URL=http://$NEW_HOST|g" .env
  sed -i.bak "s|CALLBACK_URL_SYNC=http://[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}|CALLBACK_URL_SYNC=http://$NEW_HOST|g" .env
  rm -f .env.bak
fi

# 修改docker-compose.yml文件
print_info "修改docker-compose.yml文件..."
# 如果存在Docker镜像，则使用镜像而不是构建
if [ -f "backend_image.tar" ]; then
  # 替换backend部分的build指令为image指令
  sed -i.bak "s|build: .|image: iboard_backend:latest|g" docker-compose.yml
  rm -f docker-compose.yml.bak
fi

# 步骤6: 创建部署包
print_info "步骤6: 创建部署包..."
tar -czf ../iboard_migrate.tar.gz * || handle_error "无法创建部署包"
cd ..

# 步骤7: 上传到新服务器
print_info "步骤7: 上传到新服务器..."
export SSHPASS="$NEW_PASS"
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "mkdir -p $NEW_DIR" || handle_error "无法创建目标目录"
sshpass -e scp -o StrictHostKeyChecking=no -P "$NEW_PORT" "iboard_migrate.tar.gz" "$NEW_USER@$NEW_HOST:$NEW_DIR/" || handle_error "无法上传部署包"

# 步骤8: 在新服务器上解压并准备环境
print_info "步骤8: 在新服务器上解压并准备环境..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && tar -xzf iboard_migrate.tar.gz && rm iboard_migrate.tar.gz" || handle_error "无法解压部署包"

# 如果有Docker镜像，加载它
if [ -f "temp_migrate/backend_image.tar" ]; then
  print_info "加载Docker镜像..."
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && docker load -i backend_image.tar && docker tag $BACKEND_IMAGE iboard_backend:latest && rm backend_image.tar" || print_warning "无法加载Docker镜像"
fi

# 步骤9: 准备MySQL初始化脚本
print_info "步骤9: 准备MySQL初始化脚本..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && mkdir -p mysql_init && echo 'CREATE DATABASE IF NOT EXISTS iboard_db;' > mysql_init/01-create-db.sql && echo 'USE iboard_db;' > mysql_init/02-restore-db.sql && cat iboard_db_backup.sql >> mysql_init/02-restore-db.sql" || handle_error "无法准备MySQL初始化脚本"

# 修改docker-compose.yml添加初始化脚本
print_info "修改docker-compose.yml添加初始化脚本..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && grep -q './mysql_init:/docker-entrypoint-initdb.d' docker-compose.yml || sed -i 's|mysql_data:/var/lib/mysql|mysql_data:/var/lib/mysql\\n      - ./mysql_init:/docker-entrypoint-initdb.d|' docker-compose.yml" || handle_error "无法修改docker-compose.yml"

# 步骤10: 启动服务
print_info "步骤10: 启动服务..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && docker-compose down -v || true && docker-compose up -d" || handle_error "无法启动服务"

# 步骤11: 等待服务就绪
print_info "步骤11: 等待服务就绪..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && for i in {1..60}; do if docker-compose ps | grep -q 'Up'; then echo '所有服务已就绪！'; break; fi; if [ \$i -eq 60 ]; then echo '服务启动超时'; docker-compose logs; exit 1; fi; echo '等待服务就绪... (尝试 '\$i'/60)'; sleep 5; done" || print_warning "等待服务就绪超时"

# 步骤12: 恢复Redis数据
if [ -f "temp_migrate/redis_dump.rdb" ]; then
  print_info "步骤12: 恢复Redis数据..."
  sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && docker cp redis_dump.rdb iboard_redis:/data/dump.rdb && docker-compose restart redis" || print_warning "无法恢复Redis数据"
fi

# 步骤13: 检查服务状态
print_info "步骤13: 检查服务状态..."
sshpass -e ssh -o StrictHostKeyChecking=no -p "$NEW_PORT" "$NEW_USER@$NEW_HOST" "cd $NEW_DIR && docker-compose ps" || print_warning "无法检查服务状态"

# 清理临时文件
print_info "清理临时文件..."
rm -rf temp_migrate
rm -f iboard_migrate.tar.gz

# 创建update_new.sh脚本，用于更新新服务器
print_info "创建update_new.sh脚本，用于更新新服务器..."
cat > update_new.sh << EOF
#!/bin/bash

# Configuration
REMOTE_USER="root"
REMOTE_HOST="$NEW_HOST"
REMOTE_DIR="/root/iboard"
REMOTE_PASS="$NEW_PASS"
DB_HOST="localhost"
DB_PORT="3308"
DB_USER="root"
DB_PASS="1090119your"
DB_NAME="iboard_db"

echo "Starting update process..."

# Function to handle errors
handle_error() {
  echo -e "\033[0;31mError: \$1\033[0m"
  echo "Attempting to rollback..."
    
  # Rollback command
  sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && 
    cp -r backup/services/base/notice_sync_service.go services/base/ &&
    docker-compose down && 
    docker-compose build backend && 
    docker-compose up -d"
    
  echo "Rollback completed. Please check service status."
  exit 1
}

# 1. Check service status first
echo "Checking current service status..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && docker-compose ps" || handle_error "Failed to check service status"

# 2. Create backup of current files
echo "Creating backup of current files..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && 
  mkdir -p backup/services/base &&
  cp -r services/base/notice_sync_service.go backup/services/base/ &&
  echo 'Backup created'" || handle_error "Failed to create backup"

# 3. Stop backend service gracefully
echo "Stopping backend service..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && docker-compose stop backend && sleep 5" || handle_error "Failed to stop backend service"

# 4. Create temp directory and subdirectories
if [ -d "temp" ]; then
  rm -rf "temp"
fi
mkdir -p "temp/services/base"

echo "Creating update package..."

# 5. Copy modified files
echo "Copying modified files..."
cp "services/base/notice_sync_service.go" "temp/services/base/"

# 6. Create update package
echo "Creating update archive..."
cd "temp"
tar -czf "../update.tar.gz" *
cd ..

# 7. Upload files
echo "Uploading update package..."
sshpass -p "\$REMOTE_PASS" scp -o StrictHostKeyChecking=no "update.tar.gz" "\${REMOTE_USER}@\${REMOTE_HOST}:\${REMOTE_DIR}/" || handle_error "Failed to upload update package"

# 8. Extract files and verify
echo "Extracting files and verifying..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && 
  tar -xzf update.tar.gz && 
  rm update.tar.gz && 
  ls -l services/base/notice_sync_service.go" || handle_error "Failed to extract files"

# 9. Build and start backend service
echo "Building and starting backend service..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && 
  docker-compose down && 
  docker-compose build backend && 
  docker-compose up -d" || handle_error "Failed to build and start service"

# 10. Wait for service to start
echo "Waiting for service to start..."
sleep 10

# 11. Verify service status
echo "Verifying service status..."
result=\$(sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && 
  docker-compose ps && 
  echo 'Checking service health...' && 
  curl -s http://localhost:10031/health || echo 'Service not responding'")
if [[ "\$result" == *"Service not responding"* ]]; then
  handle_error "Service failed to start properly"
fi

# 12. Clean up
echo "Cleaning up temporary files..."
rm -rf "temp"

echo "Update completed successfully!"

# 13. Check final status
echo ""
echo "Checking container status..."
sshpass -p "\$REMOTE_PASS" ssh -o StrictHostKeyChecking=no \$REMOTE_USER@\$REMOTE_HOST "cd \$REMOTE_DIR && docker-compose ps"

echo ""
echo "Important notes:"
echo "1. Backend service updated with new notice sync functionality"
echo "2. Service port remains: 10031"
echo "3. Using existing MySQL and Redis services"
echo "4. Backup of original files created in \$REMOTE_DIR/backup"
echo ""
echo "To view logs use: docker-compose logs -f backend"
echo "To restart service use: docker-compose restart backend"
echo "To rollback use: cd \$REMOTE_DIR && cp -r backup/services/base/notice_sync_service.go services/base/ && docker-compose down && docker-compose build backend && docker-compose up -d"
echo ""

read -p "Press Enter to continue..." 
EOF

chmod +x update_new.sh

print_success "迁移完成！"
print_info "本地备份文件: $BACKUP_FILE"
print_info "新服务器 ($NEW_HOST) 上的服务已经启动"
print_info "原始服务器 ($ORIGINAL_HOST) 上的服务保持不变"
print_info "已创建update_new.sh脚本，用于更新新服务器"
print_warning "注意：两个服务器的数据库是独立的，不会自动同步"

# 显示使用说明
echo ""
echo "使用说明："
echo "1. 使用 ./update.sh 更新原始服务器 ($ORIGINAL_HOST)"
echo "2. 使用 ./update_new.sh 更新新服务器 ($NEW_HOST)"
echo "3. 备份文件保存在 $BACKUP_DIR 目录中"
echo ""
echo "如果需要恢复备份，可以使用以下命令："
echo "tar -xzf $BACKUP_FILE -C /path/to/restore" 