FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/users-api ./cmd/users-api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/users-api /usr/local/bin/users-api
EXPOSE 8080
CMD ["/usr/local/bin/users-api"]
