# Configuration
$REMOTE_USER = "root"
$REMOTE_HOST = "39.108.49.167"
$REMOTE_DIR = "/root/iboard"
$REMOTE_PASS = "1090119your@"
$DB_HOST = "localhost"
$DB_PORT = "3308"
$DB_USER = "root"
$DB_PASS = "1090119your"
$DB_NAME = "iboard_db"

Write-Host "Starting update process..."

# 1. Stop all services gracefully
Write-Host "Stopping all services..."
$cmd = "cd $REMOTE_DIR && docker-compose stop && docker container prune -f"
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd
Write-Host "Waiting for services to stop..."
Start-Sleep -Seconds 5

# 2. Create temp directory
if (Test-Path "temp") {
  Remove-Item -Recurse -Force "temp"
}
New-Item -ItemType Directory -Path "temp"
Write-Host "Creating deployment package..."

# 3. Copy configuration files
Write-Host "Copying configuration files..."
Copy-Item "docker-compose.yml" -Destination "temp/" -ErrorAction SilentlyContinue
Copy-Item ".env" -Destination "temp/" -ErrorAction SilentlyContinue
Copy-Item "Dockerfile" -Destination "temp/" -ErrorAction SilentlyContinue

# 4. Copy source code files
Write-Host "Copying source code files..."
Copy-Item -Recurse "controller" -Destination "temp/"
Copy-Item -Recurse "database" -Destination "temp/"
Copy-Item -Recurse "middleware" -Destination "temp/"
Copy-Item -Recurse "models" -Destination "temp/"
Copy-Item -Recurse "router" -Destination "temp/"
Copy-Item -Recurse "services" -Destination "temp/"
Copy-Item -Recurse "utils" -Destination "temp/"
Copy-Item "go.mod" -Destination "temp/"
Copy-Item "go.sum" -Destination "temp/"
Copy-Item "main.go" -Destination "temp/"

# 5. Create deployment package
Write-Host "Creating deployment archive..."
Set-Location "temp"
tar -czf "../deploy.tar.gz" *
Set-Location ..

# 6. Upload files
Write-Host "Uploading deployment package..."
echo $REMOTE_PASS | scp -o StrictHostKeyChecking=no "deploy.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/"

# 7. Extract files
Write-Host "Extracting files..."
$cmd = "cd $REMOTE_DIR && tar -xzf deploy.tar.gz && rm deploy.tar.gz"
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

# 8. Start services and wait for MySQL to be ready
Write-Host "Starting MySQL and Redis..."
$cmd = @"
cd $REMOTE_DIR && 
docker-compose up -d mysql redis && 
echo 'Waiting for MySQL to be ready...' && 
sleep 10
"@
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

# 9. Run database migration
Write-Host "Running database migration..."
$migrationCmd = @"
SET FOREIGN_KEY_CHECKS=0;

-- Add priority to advertisements (only if not exists)
ALTER TABLE advertisements 
ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0 
COMMENT 'priority 0-100';

-- Add priority and is_ismart_notice to notices (only if not exists)
ALTER TABLE notices 
ADD COLUMN IF NOT EXISTS priority INT DEFAULT 0 
COMMENT 'priority 0-100';

ALTER TABLE notices 
ADD COLUMN IF NOT EXISTS is_ismart_notice BOOLEAN DEFAULT false 
COMMENT 'ismart notice sync flag';

-- Remove password from buildings (only if exists)
ALTER TABLE buildings 
DROP COLUMN IF EXISTS password;

-- Create devices table (only if not exists)
CREATE TABLE IF NOT EXISTS devices (
    id bigint unsigned NOT NULL AUTO_INCREMENT,
    created_at datetime(3) DEFAULT NULL,
    updated_at datetime(3) DEFAULT NULL,
    deleted_at datetime(3) DEFAULT NULL,
    building_id bigint unsigned DEFAULT NULL,
    device_id varchar(255) NOT NULL,
    device_name varchar(255) NOT NULL,
    device_type varchar(50) DEFAULT NULL,
    device_status varchar(50) DEFAULT NULL,
    last_online_time datetime(3) DEFAULT NULL,
    PRIMARY KEY (id),
    KEY idx_devices_deleted_at (deleted_at),
    KEY idx_devices_building_id (building_id),
    CONSTRAINT fk_devices_building FOREIGN KEY (building_id) REFERENCES buildings (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

SET FOREIGN_KEY_CHECKS=1;
"@
$cmd = "cd $REMOTE_DIR && echo '$migrationCmd' > migrate.sql && docker-compose exec -T mysql mysql -u root -p1090119your iboard_db < migrate.sql && rm migrate.sql"
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

# 10. Build and start backend service
Write-Host "Building and starting backend service..."
$cmd = @"
cd $REMOTE_DIR && 
docker-compose build backend && 
docker-compose up -d backend && 
echo 'Waiting for backend to start...' && 
sleep 5 && 
docker-compose logs --tail=50 backend
"@
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

# 11. Clean up
Write-Host "Cleaning up temporary files..."
Remove-Item -Recurse -Force "temp"

Write-Host "Update completed!"

# 12. Check final status
Write-Host ""
Write-Host "Checking container status..."
$cmd = "cd $REMOTE_DIR && docker-compose ps"
echo $REMOTE_PASS | ssh -o StrictHostKeyChecking=no $REMOTE_USER@$REMOTE_HOST $cmd

Write-Host ""
Write-Host "Important notes:"
Write-Host "1. Backend service updated"
Write-Host "2. Database structure updated (existing data preserved)"
Write-Host "3. Service port remains: 10031"
Write-Host "4. Using existing MySQL and Redis services"
Write-Host ""
Write-Host "To view logs use: docker-compose logs -f backend"
Write-Host "To restart service use: docker-compose restart backend"
Write-Host "To check database status use: docker-compose exec mysql mysql -u root -p1090119your -e 'SHOW TABLES;' iboard_db"
Write-Host ""

Pause 