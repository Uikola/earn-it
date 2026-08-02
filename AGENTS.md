# AGENTS.md

## Commands

```bash
task format          # gofumpt + gci (required before commit)
task lint            # golangci-lint v2
task pre-commit      # format + lint
task migrate         # apply migrations (goose)
task create-migration  # creates migrations/change_my_name.sql — rename after
task ogen-generate   # OpenAPI codegen via go generate ./...
```

Dev infrastructure: `docker compose -f deployments/docker-compose-dev.yaml up -d`  
Postgres on `:5433`, Redis on `:6379`.

## Architecture

- **Entry point**: `cmd/bot/main.go` (cmd/api is empty placeholder)
- **Telegram bot**: telebot.v3 with layout-driven UI (`telegram.yml` + `locales/ru.yml`)
- **Input state**: `intele.InputManager` backed by Redis (`internal/repository/redis`)
- **Handlers**: `internal/telegram/handlers/` — register in `bot.go:Setup()`
- **DB**: pgx/v5, connection string hardcoded in `cmd/bot/main.go:35`
- **Migrations**: `migrations/*.sql` via goose

## Conventions

- **Import order**: standard → default → `github.com/Uikola/earn-it/`
- **No `fmt.Print*`** — use `log/slog` for structured logging
- **No `time.Sleep`** in production code — use timers/context
- **No `http.DefaultClient`** — create clients with timeouts
- **Cyclop max**: 20

## Gotchas

- `.env` only contains `BOT_TOKEN`; DB/Redis credentials are hardcoded
- Migration template name is `change_my_name` — always rename after creation
- `gci` prefix in Taskfile references `github.com/Uikola/keepflame/` (stale, should be `earn-it`)
