FROM --platform=linux/amd64 golang:latest as builder
WORKDIR /app

RUN go mod download github.com/go-sql-driver/mysql@v1.7.1
RUN go mod download github.com/json-iterator/go@v1.1.12
RUN go mod download github.com/spf13/cobra@v1.8.0
RUN go mod download github.com/spf13/viper@v1.18.2

COPY . .
RUN go build -o webhook-interface


ENTRYPOINT ["/app/webhook-interface", "server"]