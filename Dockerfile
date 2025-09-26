# --- build stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app

# зависимости
COPY go.mod go.sum ./
RUN go mod download

# копируем исходники
COPY . .

# собираем бинарь
RUN go build -o travel-api ./cmd/api

# --- runtime stage ---
FROM alpine:3.20
WORKDIR /app

# certs для https-запросов
RUN apk add --no-cache ca-certificates tzdata

# копируем бинарь и миграции
COPY --from=builder /app/travel-api .
COPY migrations ./migrations

EXPOSE 8080

# по умолчанию запускаем API
CMD ["./travel-api", "serve"]
