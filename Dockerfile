#build stage
FROM golang:1.17.5-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go get -d -v ./...
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/build/server server/server.go
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/build/webserver Web/server/webserver.go
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /app/build/bot bot/bot.go

#-v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app /app
ENTRYPOINT /app
LABEL Name=patron Version=0.0.1
EXPOSE 9000
EXPOSE 8000
