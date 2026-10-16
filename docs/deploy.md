# Запуск и настройки

## Docker (VPS или локально)

```bash
cp .env.example .env    # ADMIN_PASSWORD, TELEGRAM_BOT_TOKEN, ИИ
docker compose up -d --build
```

Поднимаются два контейнера: бэкенд с файлом SQLite в volume `data` и nginx с панелью. Миграции goose применяются при старте. Бот работает через polling, публичный адрес не нужен.

> Один токен бота нельзя одновременно использовать в Docker (polling) и на Cloudflare (webhook): polling снимает вебхук.

## Cloudflare (бесплатно)

Worker на Go (wasm), D1 и Workers AI. Нужен аккаунт Cloudflare и токен с правами `Workers Scripts: Edit`, `D1: Edit`, `Workers AI: Read`.

```bash
make tools          # wrangler в .tools/
make cf-db-create   # база D1; id → CF_D1_DATABASE_ID в .env
make cf-migrate     # схема (секции Up из backend/migrations)
make cf-secrets     # ADMIN_PASSWORD, TELEGRAM_BOT_TOKEN, TELEGRAM_WEBHOOK_SECRET
make cf-deploy      # тесты, сборка, деплой; адрес → WORKER_URL в .env
make cf-webhook     # направить Telegram на Worker
```

Панель и API живут на одном адресе: статику бесплатно отдаёт Static Assets, в Go попадают только `/api/*`, `/telegram/*` и `/healthz`.

Worker собирается **TinyGo** (`make tools` ставит его в `.tools/`). Сервер и Docker собираются обычным Go, TinyGo им не нужен. `CF_COMPILER=go make cf-deploy` соберёт Worker обычным Go.

Ограничения бесплатного тарифа:

- **Бандл до 3 МБ в gzip.** На TinyGo 0.7 МБ (на обычном Go 2.2 МБ), сборку и размер проверяет CI.
- **10 мс CPU на запрос.** Холодный старт экземпляра Go в новом изоляте занимает 10–60 мс. Внутри изолята экземпляр переиспользуется (`deploy/cloudflare/worker.mjs`), но у редко используемого бота изоляты почти всегда холодные. По документации у изолята есть «встроенная гибкость» на редкие превышения, обрывают только при постоянных (ошибка 1102). На практике запросы проходят, но гарантий нет. Если начнутся 1102, поможет Workers Paid ($5, 30 с CPU) или Docker.
- **10 000 нейронов Workers AI в день** — около 2000 сообщений на Gemma 4.

| Запрос, CPU | обычный Go | TinyGo + переиспользование |
|---|---|---|
| `/healthz` | 13–235 мс | 7–57 мс |
| список трат | 27–201 мс | 37–119 мс |
| сообщение боту | 33–232 мс | 106–248 мс |

## Переменные

| Переменная | По умолчанию | |
|---|---|---|
| `ADMIN_PASSWORD` | — | пароль панели, ≥ 8 символов |
| `TELEGRAM_BOT_TOKEN` | — | от @BotFather |
| `TELEGRAM_MODE` | `polling` | `polling`, `webhook`, `off` |
| `TELEGRAM_WEBHOOK_SECRET` | — | для `webhook` |
| `LLM_PROVIDER` | `openai` | `openai`, `openrouter`, `anthropic`, `workersai` |
| `LLM_MODEL`, `LLM_BASE_URL`, `LLM_API_KEY` | — | модель, адрес и ключ |
| `LLM_EXTRA_BODY` | — | JSON-параметры модели |
| `DB_PATH` | `data/expenses.db` | файл SQLite (в Worker — D1) |
| `APP_TIMEZONE` | `Europe/Moscow` | в нём считаются «сегодня» и «вчера» |
| `BOT_MESSAGES_PER_USER_PER_DAY` | `100` | 0 — без лимита |
| `LLM_CALLS_PER_DAY` | `1000` | на весь бот |
| `BOT_MAX_MESSAGE_LENGTH` | `500` | символов |
| `ADMIN_LOGIN_ATTEMPTS_PER_HOUR` | `10` | с одного IP |

Примеры настроек ИИ (Workers AI, OpenRouter, DeepSeek, Claude, Ollama) — в `.env.example`.

## Разработка

```bash
make test      # тесты Go + проверка сборки Worker
make eval      # прогон настоящей модели по трудным фразам
make dev       # сервер локально, база в backend/data
make web-dev   # панель с hot reload, /api → make dev
```
