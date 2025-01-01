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
COPY --from=build-stage /app/.env /app/.env
COPY --from=build-stage /app/main /app

EXPOSE 10032 

ENTRYPOINT ["/app/main"]