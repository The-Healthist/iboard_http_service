## 1

```
docker volume create iboard_mysql_data1.1.2
docker volume create iboard_redis_data1.1.2

# 迁移MySQL数据
docker run --rm -v iboard_mysql_data1.1.0:/source -v iboard_mysql_data1.1.2:/dest alpine sh -c "cp -r /source/* /dest/"

# 迁移Redis数据
docker run --rm -v iboard_mysql_data1.1.0:/source -v iboard_redis_data1.1.2:/dest alpine sh -c "cp -r /source/* /dest/"


docker compose build --build-arg VERSION=1.1.2

# 2. 构建 & 推送应用镜像
docker build -t stonesea/iboard_http_service:1.1.2 .
docker push stonesea/iboard_http_service:1.1.2

# 3. 构建 & 推送 MySQL 镜像
docker build -f mysql/Dockerfile -t stonesea/iboard-mysql:1.1.2 .
docker push stonesea/iboard-mysql:1.1.2

# 4. 构建 & 推送 Redis 镜像
docker build -f redis/Dockerfile -t stonesea/iboard-redis:1.1.2 .
docker push stonesea/iboard-redis:1.1.2

# 5. 拉取并启动所有服务
docker compose pull
docker compose up -d
```