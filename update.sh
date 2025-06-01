#!/bin/bash

# Configuration
REMOTE_USER="root"
REMOTE_HOST="39.108.49.167"
REMOTE_DIR="/root/iboard"
REMOTE_PASS="1090119your@"
DB_HOST="localhost"
DB_PORT="3308"
DB_USER="root"
DB_PASS="1090119your"
DB_NAME="iboard_db"

echo "Starting update process..."

# Function to handle errors
handle_error() {
  echo -e "\033[0;31mError: $1\033[0m"
  echo "Attempting to rollback..."
    
  # Rollback command
  sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && 
    cp -r backup/services/base/notice_sync_service.go services/base/ &&
    docker-compose down && 
    docker-compose build backend && 
    docker-compose up -d"
    
  echo "Rollback completed. Please check service status."
  exit 1
}

# 1. Check service status first
echo "Checking current service status..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && docker-compose ps" || handle_error "Failed to check service status"

# 2. Create backup of current files
echo "Creating backup of current files..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && 
  mkdir -p backup/services/base &&
  cp -r services/base/notice_sync_service.go backup/services/base/ &&
  echo 'Backup created'" || handle_error "Failed to create backup"

# 3. Stop backend service gracefully
echo "Stopping backend service..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && docker-compose stop backend && sleep 5" || handle_error "Failed to stop backend service"

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
sshpass -p "$REMOTE_PASS" scp -o StrictHostKeyChecking=no "update.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/" || handle_error "Failed to upload update package"

# 8. Extract files and verify
echo "Extracting files and verifying..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && 
  tar -xzf update.tar.gz && 
  rm update.tar.gz && 
  ls -l services/base/notice_sync_service.go" || handle_error "Failed to extract files"

# 9. Build and start backend service
echo "Building and starting backend service..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && 
  docker-compose down && 
  docker-compose build backend && 
  docker-compose up -d" || handle_error "Failed to build and start service"

# 10. Wait for service to start
echo "Waiting for service to start..."
sleep 10

# 11. Verify service status
echo "Verifying service status..."
result=$(sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && 
  docker-compose ps && 
  echo 'Checking service health...' && 
  curl -s http://localhost:10031/health || echo 'Service not responding'")
if [[ "$result" == *"Service not responding"* ]]; then
  handle_error "Service failed to start properly"
fi

# 12. Clean up
echo "Cleaning up temporary files..."
rm -rf "temp"

echo "Update completed successfully!"

# 13. Check final status
echo ""
echo "Checking container status..."
sshpass -p "$REMOTE_PASS" ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST "cd $REMOTE_DIR && docker-compose ps"

echo ""
echo "Important notes:"
echo "1. Backend service updated with new notice sync functionality"
echo "2. Service port remains: 10031"
echo "3. Using existing MySQL and Redis services"
echo "4. Backup of original files created in $REMOTE_DIR/backup"
echo ""
echo "To view logs use: docker-compose logs -f backend"
echo "To restart service use: docker-compose restart backend"
echo "To rollback use: cd $REMOTE_DIR && cp -r backup/services/base/notice_sync_service.go services/base/ && docker-compose down && docker-compose build backend && docker-compose up -d"
echo ""

read -p "Press Enter to continue..." 