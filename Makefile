.PHONY: fmt test vet build run bot-run-no-gateway bot-smoke-gateway command-sync-dry-run economy-reset-dry-run economy-reset-apply-confirmed work-payout-dry-run scheduler-lease-status staging-preflight staging-command-sync-dry-run staging-command-sync-apply-guild-confirmed staging-gateway-smoke-confirmed mongo-compose-up mongo-compose-down mongo-compose-ps mongo-local-audit mongo-audit mongo-index-dry-run feature-test check

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./cmd/mhcat-bot
	go build ./cmd/mhcat-command-sync
	go build ./cmd/mhcat-mongo-audit
	go build ./cmd/mhcat-mongo-index
	go build ./cmd/mhcat-staging-preflight
	go build ./cmd/mhcat-economy-reset
	go build ./cmd/mhcat-scheduler-lease
	go build ./cmd/mhcat-work-payout

run:
	go run ./cmd/mhcat-bot

bot-run-no-gateway:
	MHCAT_DISCORD_ENABLE_GATEWAY=false go run ./cmd/mhcat-bot

bot-smoke-gateway:
	MHCAT_DISCORD_ENABLE_GATEWAY=true MHCAT_DISCORD_GATEWAY_SMOKE_TEST=true go run ./cmd/mhcat-bot

command-sync-dry-run:
	go run ./cmd/mhcat-command-sync --dry-run

economy-reset-dry-run:
	go run ./cmd/mhcat-economy-reset --dry-run

economy-reset-apply-confirmed:
	go run ./cmd/mhcat-economy-reset --apply

work-payout-dry-run:
	go run ./cmd/mhcat-work-payout --dry-run

scheduler-lease-status:
	go run ./cmd/mhcat-scheduler-lease --action status

staging-preflight:
	go run ./cmd/mhcat-staging-preflight --format text

staging-command-sync-dry-run:
	sh scripts/staging/command-sync-dry-run.sh

staging-command-sync-apply-guild-confirmed:
	sh scripts/staging/command-sync-apply-guild.sh

staging-gateway-smoke-confirmed:
	sh scripts/staging/gateway-smoke.sh

mongo-compose-up:
	docker compose up -d mongodb

mongo-compose-down:
	docker compose down

mongo-compose-ps:
	docker compose ps

mongo-local-audit:
	MHCAT_MONGODB_URI='mongodb://127.0.0.1:27018/mhcat-database?directConnection=true' MHCAT_MONGODB_DATABASE=mhcat-database go run ./cmd/mhcat-mongo-audit --format text

mongo-audit:
	go run ./cmd/mhcat-mongo-audit

mongo-index-dry-run:
	go run ./cmd/mhcat-mongo-index --dry-run

feature-test:
	go test ./internal/core/features ./internal/core/services/utility ./internal/core/services/economy ./internal/core/services/announcements ./internal/discord/features/utility ./internal/discord/features/economy ./internal/discord/features/lottery ./internal/discord/features/stats ./internal/discord/features/announcements ./internal/discord/commands ./internal/discord/interactions

check: fmt test vet build
