#!/usr/bin/env sh
# wrangler.deploy.toml с привязкой D1 и миграции для D1 (только секции Up из goose).
set -eu
cd "$(dirname "$0")"
: "${CF_D1_DATABASE_ID:?set CF_D1_DATABASE_ID (make cf-db-create prints it)}"

rm -rf d1-migrations && mkdir -p d1-migrations
for f in ../../backend/migrations/*.sql; do
  awk '/^-- \+goose Down/ {exit} !/^-- \+goose Up/' "$f" > "d1-migrations/$(basename "$f")"
done

cp wrangler.toml wrangler.deploy.toml
cat >> wrangler.deploy.toml <<TOML

[[d1_databases]]
binding = "DB"
database_name = "expense-bot"
database_id = "${CF_D1_DATABASE_ID}"
migrations_dir = "d1-migrations"
TOML
