@echo off
chcp 65001
setlocal enabledelayedexpansion

:: Configuration
set REMOTE_USER=root
set REMOTE_HOST=39.108.49.167
set REMOTE_DIR=/root/iboard_http_service_v2
set REMOTE_PASS=1090119your@
set APP_NAME=iboard_http_service
set OLD_DIR=/root/iboard/http_service

echo Starting deployment update process...

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

:: 5. Stop the old service
echo Stopping old service...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %OLD_DIR% && docker-compose down"

:: 6. Create new directory on server
echo Creating new directory on server...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "mkdir -p %REMOTE_DIR%"

:: 7. Upload the package to server
echo Uploading deployment package...
scp -o StrictHostKeyChecking=no deploy.tar.gz %REMOTE_USER%@%REMOTE_HOST%:%REMOTE_DIR%/

:: 8. Extract and update on server
echo Extracting files on server...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && tar -xzf deploy.tar.gz && rm deploy.tar.gz"

:: 9. Start the new backend service (using existing MySQL and Redis)
echo Building and starting new backend service...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && docker-compose up -d --build backend"

:: 10. Clean up local files
echo Cleaning up temporary files...
rd /s /q temp
del deploy.tar.gz

echo Deployment completed successfully!

:: 11. Check container status
echo.
echo Checking container status...
echo %REMOTE_PASS% | ssh -o StrictHostKeyChecking=no %REMOTE_USER%@%REMOTE_HOST% "cd %REMOTE_DIR% && docker-compose ps"

echo.
echo Important Notes:
echo 1. New backend is deployed to: %REMOTE_DIR%
echo 2. Using existing MySQL and Redis data
echo 3. Service ports:
echo    - Backend: 10031
echo    - MySQL: 3306 (using existing)
echo    - Redis: 6379 (using existing)
echo 4. Old service has been stopped
echo 5. To rollback, run: cd %OLD_DIR% ^&^& docker-compose up -d
echo.
echo To check logs use: docker-compose logs -f backend
echo To restart service use: docker-compose restart backend
echo.

pause 