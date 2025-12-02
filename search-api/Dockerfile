FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/search-api ./cmd/search-api

FROM alpine:3.20
WORKDIR /app

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/search-api /usr/local/bin/search-api
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/search-api"]
