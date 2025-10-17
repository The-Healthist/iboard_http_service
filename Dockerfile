# Build code
FROM golang:1.23-alpine AS build-stage

# 添加版本号作为构建参数
ARG VERSION=1.1.7

# 配置 Alpine 镜像源（加速）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY .  /app

# 构建时指定代码目录，生成 main 可执行文件
RUN go build -o main ./cmd/server

# 添加版本标签到镜像元数据
LABEL version=$VERSION

# Run release
FROM alpine:latest AS release-stage

# 从构建阶段传递版本号
ARG VERSION
LABEL version=$VERSION

WORKDIR /app
COPY --from=build-stage /app/main /app/
COPY --from=build-stage /app/.env /app/

# Create logs directory
RUN mkdir -p /app/logs && \
  chmod 755 /app/logs

EXPOSE 10031

ENTRYPOINT ["/app/main"]