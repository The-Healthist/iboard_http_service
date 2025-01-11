@echo off
chcp 65001
setlocal enabledelayedexpansion

:: Configuration
set REMOTE_USER=root
set REMOTE_HOST=39.108.49.167
set REMOTE_DIR=/root/iboard_http_service_v3
set REMOTE_PASS=1090119your@
set APP_NAME=iboard_http_service_v3
set OLD_DIR=/root/iboard/http_service

echo Starting deployment update process for v3...

:: 1. Create a temporary directory for the deployment package
if exist "temp" rd /s /q temp
mkdir temp
echo Creating deployment package...

:: 2. Copy necessary files to temp directory
echo Copying configuration files...
copy docker-compose.yml temp\ > nul
if exist .env copy .env temp\ > nul
if exist Dockerfile copy Dockerfile temp\ > nul

:: 3. Copy all source code files
echo Copying source code files...
xcopy /E /I /Y controller temp\controller\ > nul
xcopy /E /I /Y database temp\database\ > nul
xcopy /E /I /Y middleware temp\middleware\ > nul
xcopy /E /I /Y models temp\models\ > nul
xcopy /E /I /Y router temp\router\ > nul
xcopy /E /I /Y services temp\services\ > nul
xcopy /E /I /Y utils temp\utils\ > nul
copy go.mod temp\ > nul
copy go.sum temp\ > nul
copy main.go temp\ > nul

:: 4. Create tar file
echo Creating deployment archive...
cd temp
tar -czf ..\deploy.tar.gz *
cd ..

:: 5. Create migration script
echo Creating migration script...
echo #!/bin/bash > migrate_data.sh
echo echo "Starting data migration..." >> migrate_data.sh
echo. >> migrate_data.sh

:: MySQL migration
echo echo "Migrating MySQL data..." >> migrate_data.sh
echo docker exec iboard_mysql mysqldump -u root -p%DB_PASSWORD% iboard_db ^> /tmp/iboard_db_backup.sql >> migrate_data.sh
echo docker cp /tmp/iboard_db_backup.sql iboard_mysql_v3:/tmp/ >> migrate_data.sh
echo docker exec iboard_mysql_v3 /bin/sh -c "mysql -u root -p%DB_PASSWORD% iboard_db_v3 < /tmp/iboard_db_backup.sql" >> migrate_data.sh
echo rm -f /tmp/iboard_db_backup.sql >> migrate_data.sh

:: Redis migration
echo echo "Migrating Redis data..." >> migrate_data.sh
echo docker exec iboard_redis redis-cli SAVE >> migrate_data.sh
echo docker cp iboard_redis:/data/dump.rdb /tmp/redis_backup.rdb >> migrate_data.sh
echo docker cp /tmp/redis_backup.rdb iboard_redis_v3:/data/dump.rdb >> migrate_data.sh
echo docker exec iboard_redis_v3 redis-cli BGREWRITEAOF >> migrate_data.sh
echo rm -f /tmp/redis_backup.rdb >> migrate_data.sh

echo echo "Data migration completed." >> migrate_data.sh

:: 6. Create new directory on server
echo Creating new directory on server...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "mkdir -p %REMOTE_DIR%"

:: 7. Upload the package and migration script to server
echo Uploading deployment package...
scp -o StrictHostKeyChecking=no deploy.tar.gz %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/
scp -o StrictHostKeyChecking=no migrate_data.sh %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/

:: 8. Extract and update on server
echo Extracting files on server...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && tar -xzf deploy.tar.gz && rm deploy.tar.gz"

:: 9. Start the new services
echo Starting new services...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && docker-compose up -d"

:: 10. Run data migration
echo Migrating data from v1 to v3...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && chmod +x migrate_data.sh && ./migrate_data.sh"

:: 11. Clean up local files
echo Cleaning up temporary files...
rd /s /q temp
del deploy.tar.gz
del migrate_data.sh

echo Deployment completed successfully!

:: 12. Check container status
echo.
echo Checking container status...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && docker-compose ps"

echo.
echo Important Notes:
echo 1. New v3 backend is deployed to: %REMOTE_DIR%
echo 2. Service ports:
echo    - Backend: 10032
echo    - MySQL: 3309
echo    - Redis: 6380
echo 3. Original v1 service is still running
echo.
echo To check logs use: docker-compose logs -f backend
echo To restart service use: docker-compose restart backend
echo.

pause 