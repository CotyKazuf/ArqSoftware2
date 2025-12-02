FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o products-api ./cmd/products-api

FROM alpine:3.20
WORKDIR /app

RUN adduser -D -g '' appuser
USER appuser

COPY --from=builder /app/products-api /app/products-api

EXPOSE 8081
ENTRYPOINT ["/app/products-api"]
