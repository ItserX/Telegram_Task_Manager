FROM golang:1.23.2 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

COPY bot.go types.go handlers.go go.mod go.sum ./

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o server

FROM alpine:latest

COPY --from=builder /app/server ./
ENV BOT_TOKEN=570471657:AAGwNOEpO3rji0tHLVaxaVDbiQNRb6t4abk
EXPOSE 8081

CMD ["./server", "--port", "80"]