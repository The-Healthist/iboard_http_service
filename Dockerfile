# Build code
FROM golang:1.21-alpine AS build-stage

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

# Create logs directory
RUN mkdir -p /app/logs && \
  chmod 755 /app/logs

EXPOSE 10032

ENTRYPOINT ["/app/main"]