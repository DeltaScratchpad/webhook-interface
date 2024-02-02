FROM --platform=linux/amd64 golang:1-alpine as builder
WORKDIR /app

COPY . .
RUN go mod download && go mod verify
RUN go build -v -o webhook-interface

FROM --platform=linux/amd64 alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/webhook-interface /app/webhook-interface


ENTRYPOINT ["/app/webhook-interface"]