# Образ для облака: телефонная книга (ПР5) = Angular-клиент + сервер на Go.
# Сборка в три этапа (multi-stage): в итоговый образ попадают только
# готовый бинарник и собранные статические файлы, без Node и Go.

# ---------- 1. Собираем Angular ----------
FROM node:22-alpine AS client
WORKDIR /client
# сначала только package*.json: если зависимости не менялись, Docker возьмёт npm install из кэша
COPY client/package*.json ./
RUN npm install --legacy-peer-deps
COPY client/ ./
RUN npx ng build                     # результат: /client/dist/client/browser

# ---------- 2. Собираем сервер на Go ----------
FROM golang:1.25-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download                  # зависимости кэшируются отдельным слоем
COPY *.go ./
# CGO_ENABLED=0 — статический бинарник, которому не нужны системные библиотеки
RUN CGO_ENABLED=0 go build -o /phonebook .

# ---------- 3. Итоговый маленький образ ----------
FROM alpine:3.20
# корневые сертификаты нужны для TLS-подключения к MongoDB Atlas
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=server /phonebook ./phonebook
COPY --from=client /client/dist/client/browser ./static
# где лежит собранный Angular внутри контейнера
ENV STATIC_DIR=/app/static
# PORT задаёт облако (Render передаёт свой), 8080 — значение по умолчанию в main.go
EXPOSE 8080
CMD ["./phonebook"]
