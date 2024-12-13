FROM golang:1.23.0-alpine

WORKDIR /app

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["./main"] 