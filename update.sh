#!/bin/bash

# Configuration
REMOTE_USER="root"
REMOTE_HOST="39.108.49.167"
REMOTE_DIR="/root/iboard/http_service"
SUPERVISOR_NAME="iboard_http_service"
APP_NAME="iboard_http_service"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting deployment update process...${NC}"

# 1. Build for Linux
echo "Building for Linux..."
export GOOS=linux
export GOARCH=amd64
go build -o $APP_NAME
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}Build successful${NC}"

# 2. Create backup directory on remote server
echo "Creating backup directory..."
ssh $REMOTE_USER@$REMOTE_HOST "mkdir -p ${REMOTE_DIR}/backups"

# 3. Backup current version
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
echo "Backing up current version..."
ssh $REMOTE_USER@$REMOTE_HOST "if [ -f ${REMOTE_DIR}/${APP_NAME} ]; then cp ${REMOTE_DIR}/${APP_NAME} ${REMOTE_DIR}/backups/${APP_NAME}_${TIMESTAMP}; fi"

# 4. Stop the service
echo "Stopping the service..."
ssh $REMOTE_USER@$REMOTE_HOST "supervisorctl stop ${SUPERVISOR_NAME}"

# 5. Upload new binary
echo "Uploading new binary..."
scp $APP_NAME $REMOTE_USER@$REMOTE_HOST:${REMOTE_DIR}/__temp_${APP_NAME}
if [ $? -ne 0 ]; then
    echo -e "${RED}Upload failed${NC}"
    exit 1
fi

# 6. Replace old binary with new one
echo "Replacing binary..."
ssh $REMOTE_USER@$REMOTE_HOST "mv ${REMOTE_DIR}/__temp_${APP_NAME} ${REMOTE_DIR}/${APP_NAME} && chmod +x ${REMOTE_DIR}/${APP_NAME}"

# 7. Upload docker-compose.yml if changed
echo "Uploading docker-compose.yml..."
scp docker-compose.yml $REMOTE_USER@$REMOTE_HOST:${REMOTE_DIR}/docker-compose.yml

# 8. Upload .env file if it exists
if [ -f .env ]; then
    echo "Uploading .env file..."
    scp .env $REMOTE_USER@$REMOTE_HOST:${REMOTE_DIR}/.env
fi

# 9. Restart the service
echo "Restarting the service..."
ssh $REMOTE_USER@$REMOTE_HOST "supervisorctl start ${SUPERVISOR_NAME}"

# 10. Clean up local build
echo "Cleaning up..."
rm $APP_NAME

# 11. Verify service status
echo "Verifying service status..."
ssh $REMOTE_USER@$REMOTE_HOST "supervisorctl status ${SUPERVISOR_NAME}"

echo -e "${GREEN}Deployment update completed successfully!${NC}" 