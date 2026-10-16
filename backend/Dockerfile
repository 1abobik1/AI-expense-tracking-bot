FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server && mkdir -p /out/data

# distroless: только бинарник и CA-сертификаты для HTTPS к Telegram и LLM.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
# Владелец nonroot: Docker переносит права в новый volume, иначе SQLite не создаст файл.
COPY --from=build --chown=nonroot:nonroot /out/data /data
ENV DB_PATH=/data/expenses.db
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/server"]
