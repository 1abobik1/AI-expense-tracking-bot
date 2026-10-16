# REST API панели

Все запросы, кроме входа, требуют `Authorization: Bearer <token>`. Суммы передаются в рублях, даты — `YYYY-MM-DD`.

| | |
|---|---|
| `POST /api/login` | `{"password"}` → `{"token","expires_at"}` |
| `GET /api/expenses?user_id=&from=&to=&category=&limit=&offset=` | `{"items":[…],"total":{"count","amount","by_category"}}` |
| `POST /api/expenses` | `{"user_id","amount","category","spent_at","comment"}`, только одобренным участникам |
| `GET / PATCH / DELETE /api/expenses/{id}` | PATCH меняет только переданные поля |
| `GET /api/members` | участники со статусом `pending`, `approved`, `rejected` |
| `PATCH /api/members/{id}` | `{"status","name"}`: одобрить, отклонить, переименовать |
| `GET /api/categories` | категории для фильтров |
| `POST /telegram/webhook` | апдейты Telegram, проверяется секретный заголовок |
| `GET /healthz` | `ok` |

Ошибки возвращаются как `{"error": "..."}`: 401 — без входа, 404 — не найдено, 422 — ошибка валидации (с полем `field`), 429 — слишком много попыток входа.
