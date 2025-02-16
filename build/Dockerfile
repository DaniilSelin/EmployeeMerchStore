FROM golang:1.22 AS builder

WORKDIR /app

# Копируем модули и зависимости
COPY ../go.mod ../go.sum ./
RUN go mod download

# Копируем весь исходный код (учитывая, что мы в ./build)
COPY ../ ./

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o merch-system ./cmd/server/main.go

# Используем минимальный образ для финального контейнера
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

# Копируем собранный бинарник
COPY --from=builder /app/merch-system .

# Копируем конфигурацию
COPY ../config/config.yml ./config/config.yml

# Новая строка: 
COPY internal/database/migrations ./internal/database/migrations

EXPOSE 8080

CMD ["./merch-system"]
