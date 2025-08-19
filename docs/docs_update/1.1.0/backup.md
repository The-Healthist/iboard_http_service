# 1 备份mysql

``` 
# 1.1 (可选) 设置 MySQL 密码环境变量，避免 -p 引号麻烦
export MYSQL_PWD='1090119your@'

# 1.2 导出所有数据库到本地文件
docker exec -i iboard_mysql \
  mysqldump -u root --all-databases \
  > iboard_db_backup.sql

# 1.3 触发 Redis 持久化并拷贝 RDB
docker exec iboard_redis redis-cli SAVE
docker cp iboard_redis:/data/dump.rdb ./redis_dump.rdb

# 1.4 确认文件存在
ls -lh iboard_db_backup.sql redis_dump.rdb
```

```
scp -P 22 root@39.108.49.167:~/iboard/iboard_db_backup.sql .\
scp -P 22 root@39.108.49.167:~/iboard/redis_dump.rdb .\
```


# 2 重新部署脚本
```

# 1. 拷贝 SQL 到容器
docker cp .\iboard_db_backup.sql iboard_mysql:/tmp/iboard_db_backup.sql

# 2. 在容器里以 UTF-8 导入（保留所有中文和特殊字符）
docker exec iboard_mysql sh -c `
  "mysql --default-character-set=utf8mb4 -uroot -p1090119your iboard_db < /tmp/iboard_db_backup.sql"

# 3. 验证表已恢复
docker exec iboard_mysql mysql -uroot -p1090119your `
  -e "USE iboard_db; SHOW TABLES;"


# 清空旧数据
docker exec iboard_redis redis-cli FLUSHALL

# 覆盖 RDB 文件
docker cp .\redis_dump.rdb iboard_redis:/data/dump.rdb

# 重启让 RDB 生效
docker restart iboard_redis

# 验证
docker exec iboard_redis redis-cli PING
```