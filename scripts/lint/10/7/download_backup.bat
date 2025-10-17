@echo off
REM iBoard 备份下载脚本 (Windows)
REM 在本地 Windows 机器上执行

set SERVER_IP=39.108.49.167
set BACKUP_DATE=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%
set LOCAL_BACKUP_DIR=C:\Users\%USERNAME%\Documents\iboard_backup_%BACKUP_DATE%

echo 🔄 开始下载 iBoard 备份...
echo 服务器IP: %SERVER_IP%
echo 本地目录: %LOCAL_BACKUP_DIR%

REM 创建本地备份目录
mkdir "%LOCAL_BACKUP_DIR%"
cd /d "%LOCAL_BACKUP_DIR%"

echo 📥 下载备份文件...
REM 下载所有备份文件
scp -r root@%SERVER_IP%:~/backup_*/* ./

echo ✅ 下载完成！
echo 📁 备份位置: %LOCAL_BACKUP_DIR%

dir

echo.
echo 📋 下一步: 重置服务器后使用 deploy_iboard.sh 脚本部署
pause
