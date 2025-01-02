# Build code
FROM golang:1.23.0-alpine AS build-stage

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY .  /app

RUN go build -o main .

# Run release
FROM alpine:3.14 AS release-stage

WORKDIR /app
COPY --from=build-stage /app/main /app/
COPY --from=build-stage /app/.env /app/

# 添加健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:10031/health || exit 1

EXPOSE 10031

ENTRYPOINT ["/app/main"]