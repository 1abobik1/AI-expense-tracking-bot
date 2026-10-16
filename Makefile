SHELL := /bin/bash
ENV := set -a; [ -f .env ] && . ./.env; set +a
WRANGLER_VERSION := 4.135.0
LOCAL_WRANGLER := $(CURDIR)/.tools/node_modules/.bin/wrangler
WRANGLER := $$( [ -x $(LOCAL_WRANGLER) ] && echo $(LOCAL_WRANGLER) || echo "npx --yes wrangler@$(WRANGLER_VERSION)" )
# Worker собирается TinyGo (бандл втрое меньше); CF_COMPILER=go — обычным Go. Сервер TinyGo не нужен.
TINYGO_VERSION := 0.42.0
LINT_VERSION := v2.14.0
LINT := $(CURDIR)/.tools/bin/golangci-lint
LEFTHOOK_VERSION := v2.1.15
TINYGO := $(CURDIR)/.tools/tinygo/bin/tinygo
CF_COMPILER ?= tinygo
CF := deploy/cloudflare
CF_CONFIG := wrangler.deploy.toml

.DEFAULT_GOAL := help
.PHONY: help test lint fmt hooks up down logs dev web-dev web-build \
	tools eval cf-build cf-dev cf-db-create cf-migrate cf-secrets cf-deploy cf-webhook cf-tail clean

help: ## показать это меню
	@grep -hE '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*## /\t/' | expand -t16


# ── Разработка ────────────────────────────────────────────────────────────────

test: ## тесты Go и проверка сборки Worker
	cd backend && go test ./... && GOOS=js GOARCH=wasm go vet ./cmd/worker ./internal/agent/clients/llm

lint: ## golangci-lint (backend/.golangci.yml) и prettier для панели
	cd backend && $(LINT) run ./... && GOOS=js GOARCH=wasm $(LINT) run ./cmd/worker ./internal/agent/clients/llm
	cd web && npm run lint

fmt: ## отформатировать код бэкенда и панели
	cd backend && $(LINT) fmt ./... && GOOS=js GOARCH=wasm $(LINT) fmt ./cmd/worker ./internal/agent/clients/llm
	cd web && npm run format

eval: ## прогон ИИ по трудным фразам (модель из .env, можно LLM_EVAL_MODELS=a,b)
	@$(ENV); cd backend && LLM_EVAL=1 go test -count=1 -timeout 30m ./internal/agent/usecase -run TestEval -v | grep -E "==|FAIL"

dev: ## сервер локально (настройки из .env, база в backend/data); панель — make web-dev
	@$(ENV); cd backend && DB_PATH=data/dev.db HTTP_ADDR=:8080 go run ./cmd/server

web-dev: ## панель с hot reload на http://localhost:5173 (API — make dev)
	cd web && npm install --no-audit --no-fund && npm run dev

web-build: ## собрать панель в web/dist
	cd web && npm ci --no-audit --no-fund && npm run build

# ── Вариант «аренда»: docker compose ──────────────────────────────────────────

up: ## поднять всё: бэкенд и панель (http://localhost:$WEB_PORT)
	docker compose up -d --build

down: ## остановить (база остаётся в volume)
	docker compose down

logs: ## логи бэкенда
	docker compose logs -f app

# ── Вариант «бесплатно»: Cloudflare ───────────────────────────────────────────

tools: ## поставить wrangler, TinyGo, golangci-lint и lefthook в .tools/ (один раз; глобально ничего не ставится)
	@mkdir -p .tools && echo '{"name":"tools","private":true}' > .tools/package.json && \
		cd .tools && npm_config_prefix=$(CURDIR)/.tools npm i --no-audit --no-fund wrangler@$(WRANGLER_VERSION) >/dev/null && \
		echo "wrangler $$(node_modules/.bin/wrangler --version | tail -1)"
	@[ -x $(TINYGO) ] || curl -sL https://github.com/tinygo-org/tinygo/releases/download/v$(TINYGO_VERSION)/tinygo$(TINYGO_VERSION).linux-amd64.tar.gz | tar -xz -C .tools
	@$(TINYGO) version
	@[ -x $(LINT) ] || GOBIN=$(CURDIR)/.tools/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)
	@$(LINT) version | head -1
	@[ -x .tools/bin/lefthook ] || GOBIN=$(CURDIR)/.tools/bin go install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)

hooks: ## pre-commit: линтер и форматирование только для файлов коммита (.lefthook.yml)
	.tools/bin/lefthook install

cf-build: web-build ## собрать Worker (TinyGo → wasm) и панель
ifeq ($(CF_COMPILER),tinygo)
	@[ -x $(TINYGO) ] || { echo "нет TinyGo — запусти make tools"; exit 1; }
	cd backend && go run github.com/syumai/workers-go/cmd/workers-assets-gen@v0.36.0 -mode=tinygo -o ../$(CF)/build
	cd backend && $(TINYGO) build -o ../$(CF)/build/app.wasm -target wasm -no-debug ./cmd/worker
else
	cd backend && go run github.com/syumai/workers-go/cmd/workers-assets-gen@v0.36.0 -mode=go -o ../$(CF)/build
	cd backend && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ../$(CF)/build/app.wasm ./cmd/worker
endif
	@printf 'worker.wasm gzip: %s KiB (лимит бесплатного тарифа — 3072)\n' $$(( $$(gzip -9 -c $(CF)/build/app.wasm | wc -c) / 1024 ))

cf-db-create: ## создать базу D1 (один раз); id положи в .env как CF_D1_DATABASE_ID
	cd $(CF) && $(WRANGLER) d1 create expense-bot

cf-migrate: ## применить миграции к D1 в Cloudflare
	@$(ENV); $(CF)/build_config.sh && cd $(CF) && $(WRANGLER) d1 migrations apply expense-bot --remote -c $(CF_CONFIG)

cf-secrets: ## записать секреты Worker из .env
	@$(ENV); $(CF)/build_config.sh && cd $(CF) && for name in ADMIN_PASSWORD TELEGRAM_BOT_TOKEN TELEGRAM_WEBHOOK_SECRET; do \
		printenv $$name | $(WRANGLER) secret put $$name -c $(CF_CONFIG); done

cf-deploy: test cf-build ## тесты, сборка, деплой Worker и проверка /healthz
	@$(ENV); $(CF)/build_config.sh && cd $(CF) && $(WRANGLER) deploy -c $(CF_CONFIG)
	@$(ENV); sleep 3; printf 'health: '; curl -s -w ' [%{http_code}]\n' "$$WORKER_URL/healthz"

cf-webhook: ## направить вебхук Telegram на Worker
	@$(ENV); curl -s "https://api.telegram.org/bot$$TELEGRAM_BOT_TOKEN/setWebhook" \
		--data-urlencode "url=$$WORKER_URL/telegram/webhook" \
		--data-urlencode "secret_token=$$TELEGRAM_WEBHOOK_SECRET" \
		--data-urlencode 'allowed_updates=["message"]' \
		--data-urlencode "drop_pending_updates=true"; echo

cf-dev: cf-build ## Worker локально (D1 эмулируется; нужен deploy/cloudflare/.dev.vars)
	@[ -f $(CF)/.dev.vars ] || { echo "нет $(CF)/.dev.vars — скопируй .dev.vars.example"; exit 1; }
	@cd $(CF) && CF_D1_DATABASE_ID=local ./build_config.sh && \
		$(WRANGLER) d1 migrations apply expense-bot --local -c $(CF_CONFIG) && \
		$(WRANGLER) dev -c $(CF_CONFIG)

cf-tail: ## живые логи Worker
	@$(ENV); $(CF)/build_config.sh && cd $(CF) && $(WRANGLER) tail -c $(CF_CONFIG) --format pretty

clean: ## удалить артефакты сборки
	rm -rf $(CF)/build $(CF)/d1-migrations $(CF)/$(CF_CONFIG) $(CF)/.wrangler web/dist backend/data
