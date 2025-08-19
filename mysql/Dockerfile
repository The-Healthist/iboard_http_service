# MySQL Dockerfile: initialize database with backup
FROM mysql:8.0

# 设置数据库环境变量
ENV MYSQL_ROOT_PASSWORD=1090119your
ENV MYSQL_DATABASE=iboard_db

# 拷贝备份脚本以在初始化时导入数据
COPY iboard_db_backup.sql /docker-entrypoint-initdb.d/
