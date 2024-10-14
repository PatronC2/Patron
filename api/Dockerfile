# Stage 1: Build the Go API
FROM golang:1.22.3-alpine AS build

WORKDIR /app
COPY . ./
RUN go mod download
RUN go build -o api ./api.go

FROM alpine:latest
ENV GIN_MODE=release
COPY .env ./
COPY --from=build /app/api /app/api
WORKDIR /app
EXPOSE 8000

CMD ["./api"]
