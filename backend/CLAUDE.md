# backend/CLAUDE.md

Go 1.26, модуль `expensebot`. Структура по образцу service-mani: домены в `internal/<домен>`, слои внутри домена.

## Карта

```
cmd/server, cmd/worker   точки входа (VPS и Cloudflare); общая сборка — server/app.go
config/                  env → Config: caarlos0/env + go-playground/validator, только теги
migrations/              goose, SQLite
internal/app/            общее: apperr, clock, http (Router, Doer, JSON), sqlite (+conn), telegram
internal/<домен>/        expenses · members · agent · limits · auth
```

Домен устроен так (пример — `expenses`):

```
errors.go       конструкторы ошибок: apperr.ErrNotFound / ErrValidation
repository.go   интерфейсы репозиториев
usecase.go      интерфейсы юзкейсов
clients.go      интерфейсы внешних клиентов (LLM, Telegram), если есть
models/         модели и их валидация
usecase/        реализация
repository/sqlite/<сущность>/  entity.go · queries.go · mappers.go · <сущность>.go
transport/http/admin/          handlers.go (маршруты) · dto.go · mappers.go · <ресурс>.go
```

## Правила

- **Интерфейсы** — только в `repository.go`, `usecase.go`, `clients.go` домена или в `internal/app`. В обработчиках и реализациях их не объявлять.
- **Зависимости** — через интерфейсы, конкретные типы собираются только в `server/app.go`.
- **Мапперы на границах:** entity ↔ модель в репозитории, DTO ↔ модель в транспорте. Модели не получают json-тегов.
- **Ошибки** — `apperr` с кодом. Транспорт выбирает HTTP-статус через `apphttp.Fail`, агент показывает `Message` пользователю.
- **Траты всегда в `models.Scope`.** В SQL он превращается только через `whereScope`. Новый запрос к тратам начинать с него. `user_id` не брать из текста или от ИИ.
- **Совместимость с Worker (TinyGo, wasm):**
  - маршруты регистрировать через `apphttp.Router`, а не `http.ServeMux`;
  - HTTP-клиенты принимать как `apphttp.Doer`, а не `*http.Client`;
  - драйвер SQLite и goose импортирует только `internal/app/sqlite/conn`; пакеты, нужные Worker, их не тянут;
  - тяжёлую инициализацию (регулярки и т. п.) делать лениво (`sync.OnceValue`): Worker часто стартует с нуля.
- **Новая настройка** — поле с тегами `env`, `envDefault`, `validate` в `config/config.go`; в типы доменов её переводят мапперы `server/mappers.go`. Ручного разбора env нет.
- **Время** — через `clock.Clock`, не `time.Now()` в юзкейсах: тесты подменяют часы.

## Как добавить фичу

1. Схема: новый файл `migrations/000N_*.sql` с секциями `-- +goose Up` / `-- +goose Down`. Только SQLite-совместимый SQL. Уже применённые файлы не менять.
2. Модель и валидация — в `models/`, ошибки — в `errors.go`.
3. Интерфейс — в `repository.go`, реализация — в `repository/sqlite/...` (entity, queries, mappers).
4. Интерфейс юзкейса — в `usecase.go`, реализация — в `usecase/`.
5. Транспорт: DTO и мапперы, затем маршрут в `Register`.
6. Подключить в `server/app.go`.
7. Тесты: репозиторий — на `testutil.SQLite(t)`; сквозной сценарий — в `server/app_test.go`.

## ИИ-агент (`internal/agent`)

Модель возвращает только JSON-`Intent` (`usecase/intent.go`). Даты, поиск траты и ответы считает код. Сленг в суммах переводится в цифры до модели (`usecase/amounts.go`, тест `amounts_test.go`). После правок промпта или нормализатора — `make eval`, все кейсы должны пройти. Новые трудные фразы добавлять в `usecase/eval_test.go`.

## Команды

```bash
make test   # go test + vet сборки Worker
make lint   # golangci-lint (.golangci.yml) для сервера и wasm
make fmt    # goimports + gofmt; пустые строки вокруг блоков (wsl_v5, nlreturn) — make lint с --fix
make eval   # настоящая модель по трудным фразам
```
