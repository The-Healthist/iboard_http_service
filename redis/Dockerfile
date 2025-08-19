# Redis Dockerfile: include existing dump.rdb
FROM redis:7.4.1

# 拷贝持久化数据文件
COPY redis_dump.rdb /data/dump.rdb

# 设置数据目录权限
VOLUME /data
