# Build code
FROM golang:1.23.0-alpine AS build-stage

# 添加版本号作为构建参数
ARG VERSION=1.0.0

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY .  /app

# 构建时添加版本号标签
RUN go build -o main .

# 添加版本标签到镜像元数据
LABEL version=$VERSION

# Run release
FROM alpine:3.14 AS release-stage

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