COMPOSE := docker compose
EXEC := docker exec -it
LOGS := docker logs
MIGRATE_CMD ?= up
SEED_CMD ?= up
STORAGE_PROVIDER ?= seaweedfs

PROJECT_NAME := collabsphere
DEPLOY_DIR := deploy
ENV_FILE := --env-file $(DEPLOY_DIR)/.env.dev

INFRA_FILE := docker-compose.postgres.yaml
PLATFORM_FILE := docker-compose.platform.yaml
STORAGE_FILE := docker-compose.storage.$(STORAGE_PROVIDER).yaml
MIGRATE_FILE := docker-compose.migrate.yaml

# посмотри в .env какой EXTERNAL_NETWORK_NAME
NETWORK_NAME := external.network

LOG_DIR := ./logs/docker
APP_LOG := $(LOG_DIR)/app.log
MIGRATE_LOG := $(LOG_DIR)/migrate.log
SEED_LOG := $(LOG_DIR)/seed.log
CODEBASE_LOG := $(LOG_DIR)/codebase.log

CODEBASE_OUTPUT := ./docs/codebase_actual.md

APPLICATION_PORT :=

COMPOSE_ARGS = \
	$(ENV_FILE) \
	-f $(DEPLOY_DIR)/$(INFRA_FILE) \
	-f $(DEPLOY_DIR)/$(PLATFORM_FILE) \
	-f $(DEPLOY_DIR)/$(STORAGE_FILE) \
	--profile local

SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

API_HEALTH_URL ?= http://localhost:8080/health
WAIT_TIMEOUT_SEC ?= 60

MIGRATE_COMPOSE_ARGS = \
	$(ENV_FILE) \
	-p collabsphere-migrate \
	-f deploy/docker-compose.migrate.yaml \
	--profile migrate

SEED_COMPOSE_ARGS = \
	$(ENV_FILE) \
	-p collabsphere-seed \
	-f deploy/docker-compose.migrate.yaml \
	--profile seed

.PHONY: collabsphere-init network up-app up-dev up-prod down logs sync migrate seed clean-logs migrations-build check-migrations seed-reset-demo

collabsphere-init: network up-dev migrate

network:
	@docker network create $(NETWORK_NAME) >/dev/null 2>&1 || true

clean-logs:
	@mkdir -p $(LOG_DIR)
	@rm -f $(RUN_LOG)

up-dev: clean-logs
	@mkdir -p $(dir $(APP_LOG))
	@echo "Running application... (log: $(APP_LOG))"

	@if command -v gum >/dev/null 2>&1; then \
		gum spin --spinner dot --title "docker compose up (build+recreate)..." -- \
			bash -lc '$(COMPOSE) $(COMPOSE_ARGS) up -d --build --force-recreate >"$(APP_LOG)" 2>&1'; \
	else \
		echo "gum not found: running without spinner (install: go install github.com/charmbracelet/gum@latest)"; \
		$(COMPOSE) $(COMPOSE_ARGS) up -d --build --force-recreate >"$(APP_LOG)" 2>&1; \
	fi

	@if [ -n "$(API_HEALTH_URL)" ]; then \
		if command -v gum >/dev/null 2>&1; then \
			gum spin --spinner dot --title "waiting for API $(API_HEALTH_URL)..." -- \
				bash -lc '\
					deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
					while [ $$SECONDS -lt $$deadline ]; do \
						if command -v curl >/dev/null 2>&1 && curl -fsS "$(API_HEALTH_URL)" >/dev/null; then exit 0; fi; \
						sleep 0.5; \
					done; \
					exit 1'; \
		else \
			deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
			while [ $$SECONDS -lt $$deadline ]; do \
				if command -v curl >/dev/null 2>&1 && curl -fsS "$(API_HEALTH_URL)" >/dev/null; then break; fi; \
				sleep 0.5; \
			done; \
		fi; \
	fi \
	|| (echo "API not ready. Last log lines:"; tail -n 160 "$(APP_LOG)" || true; exit 1)

	@echo "✓ Started"
	@$(COMPOSE) $(COMPOSE_ARGS) ps

migrate:
	@mkdir -p $(dir $(MIGRATE_LOG))
	@docker compose $(MIGRATE_COMPOSE_ARGS) \
		run --rm --build -e MIGRATE_CMD=$(MIGRATE_CMD) migrate \
		> $(MIGRATE_LOG) 2>&1 \
	|| (echo "migrate failed; tail:"; tail -n 160 $(MIGRATE_LOG) || true; exit 1)

seed:
	@mkdir -p $(dir $(SEED_LOG))
	@docker compose $(SEED_COMPOSE_ARGS) \
		run --rm --build -e SEED_CMD=$(SEED_CMD) seed \
		> $(SEED_LOG) 2>&1 \
	|| (echo "seed failed; tail:"; tail -n 160 $(SEED_LOG) || true; exit 1)

migrations-build:
	go -C platform run ./internal/runtime/infrastructure/db/cmd/build-migrations

check-migrations:
	./scripts/build-migrations.sh
	git diff --exit-code -- platform/internal/runtime/infrastructure/db/migrations

migrate-up:
	@$(MAKE) migrate MIGRATE_CMD=up

migrate-down:
	@$(MAKE) migrate MIGRATE_CMD=down

seed-up:
	@$(MAKE) seed SEED_CMD=up
seed-status:
	@$(MAKE) seed SEED_CMD=status

seed-reset:
	@$(MAKE) seed SEED_CMD=reset

seed-reset-demo:
	@$(MAKE) seed SEED_CMD=reset-demo

codebase:
	@mkdir -p $(LOG_DIR)
	@mkdir -p ./docs
	@echo Generating codebase markdown...
	@rm -f $(CODEBASE_OUTPUT)
	@codeweaver -input=. -output=$(CODEBASE_OUTPUT) \
		-include="\\.go$$, \\.mod$$, \\.md$$,\\.sql$$,\\.yaml$$" \
		-ignore="^\\.git,^docs/" \
		> $(CODEBASE_LOG) 2>&1

logs:
	$(LOGS) -f $(APP_CONTAINER)
