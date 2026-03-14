COMPOSE := docker compose
EXEC := docker exec -it
LOGS := docker logs
MIGRATE_CMD ?= up
SEED_CMD ?= up
STORAGE_PROVIDER ?= seaweedfs
IMAGE_TAG ?= dev

PROJECT_NAME := collabsphere
DEPLOY_DIR := deploy
COMPOSE_DIR := $(DEPLOY_DIR)/compose
BASE_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.dev
POSTGRES_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.postgres.dev
STORAGE_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.storage.dev
REDIS_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.redis.dev
ZITADEL_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.zitadel.dev
WEB_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.web.dev
WEB_LOGIN_ENV_FILE := --env-file $(DEPLOY_DIR)/env/.env.web.login.dev

CORE_FILE := core.yaml
STORAGE_FILE := storage.yaml
AUTH_FILE := auth.yaml
JOBS_FILE := jobs.yaml

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
	$(BASE_ENV_FILE) \
	$(POSTGRES_ENV_FILE) \
	$(STORAGE_ENV_FILE) \
	$(REDIS_ENV_FILE) \
	$(ZITADEL_ENV_FILE) \
	-f $(COMPOSE_DIR)/$(CORE_FILE) \
	-f $(COMPOSE_DIR)/$(STORAGE_FILE) \
	-f $(COMPOSE_DIR)/$(AUTH_FILE) \
	--profile local

COMPOSE_ARGS_WITH_WEB = \
	$(BASE_ENV_FILE) \
	$(POSTGRES_ENV_FILE) \
	$(STORAGE_ENV_FILE) \
	$(REDIS_ENV_FILE) \
	$(ZITADEL_ENV_FILE) \
	$(WEB_ENV_FILE) \
	$(WEB_LOGIN_ENV_FILE) \
	-f $(COMPOSE_DIR)/$(CORE_FILE) \
	-f $(COMPOSE_DIR)/$(STORAGE_FILE) \
	-f $(COMPOSE_DIR)/$(AUTH_FILE) \
	--profile local \
	--profile web

SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

API_HEALTH_URL ?= http://localhost:8080/health
WEB_HEALTH_URL ?= http://localhost:3002/
WEB_LOGIN_HEALTH_URL ?= http://localhost:3000/ui/v2/login/login
WAIT_TIMEOUT_SEC ?= 60

MIGRATE_COMPOSE_ARGS = \
	$(BASE_ENV_FILE) \
	$(POSTGRES_ENV_FILE) \
	-p collabsphere-migrate \
	-f $(COMPOSE_DIR)/$(JOBS_FILE) \
	--profile migrate

SEED_COMPOSE_ARGS = \
	$(BASE_ENV_FILE) \
	$(POSTGRES_ENV_FILE) \
	-p collabsphere-seed \
	-f $(COMPOSE_DIR)/$(JOBS_FILE) \
	--profile seed

CONTRACTS_ENV = \
	APPLICATION_TITLE=CollabSphere \
	APPLICATION_VERSION=dev \
	APPLICATION_ENVIRONMENT=dev \
	APPLICATION_LOG_LEVEL=ERROR

.PHONY: collabsphere-init network up-app up-dev up-dev-web up-prod down logs sync migrate seed clean-logs migrations-build check-migrations check-contracts contracts-openapi-json contracts-openapi-yaml contracts-routes contracts-snapshot seed-reset-demo test-accounts-integration test-organizations-integration test-platform-reviews-integration smoke-account-signup-login smoke-bootstrap smoke-auth-legacy smoke-auth-zitadel-e2e platform-image-build web-image-build

collabsphere-init: network up-dev migrate

network:
	@docker network create $(NETWORK_NAME) >/dev/null 2>&1 || true

clean-logs:
	@mkdir -p $(LOG_DIR)
	@rm -f $(RUN_LOG)

up-dev: clean-logs
	@mkdir -p $(dir $(APP_LOG))
	@echo "Running application... (log: $(APP_LOG))"
	@echo "Deploy progress:"
	@$(COMPOSE) $(COMPOSE_ARGS) up -d --build --force-recreate 2>&1 | tee "$(APP_LOG)" \
	|| (echo "compose up failed. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS) ps || true; exit 1)

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
	|| (echo "API not ready. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS) ps || true; echo "Last log lines:"; tail -n 160 "$(APP_LOG)" || true; exit 1)

	@echo "✓ Started"
	@$(COMPOSE) $(COMPOSE_ARGS) ps

up-dev-web: clean-logs
	@mkdir -p $(dir $(APP_LOG))
	@echo "Running full platform with web surfaces... (log: $(APP_LOG))"
	@echo "Deploy progress:"
	@$(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) up -d --build --force-recreate --remove-orphans 2>&1 | tee "$(APP_LOG)" \
	|| (echo "compose up failed. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) ps || true; exit 1)

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
	|| (echo "API not ready. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) ps || true; echo "Last log lines:"; tail -n 160 "$(APP_LOG)" || true; exit 1)

	@if [ -n "$(WEB_HEALTH_URL)" ]; then \
		if command -v gum >/dev/null 2>&1; then \
			gum spin --spinner dot --title "waiting for web $(WEB_HEALTH_URL)..." -- \
				bash -lc '\
					deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
					while [ $$SECONDS -lt $$deadline ]; do \
						if command -v curl >/dev/null 2>&1 && curl -fsS "$(WEB_HEALTH_URL)" >/dev/null; then exit 0; fi; \
						sleep 0.5; \
					done; \
					exit 1'; \
		else \
			deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
			while [ $$SECONDS -lt $$deadline ]; do \
				if command -v curl >/dev/null 2>&1 && curl -fsS "$(WEB_HEALTH_URL)" >/dev/null; then break; fi; \
				sleep 0.5; \
			done; \
		fi; \
	fi \
	|| (echo "Web not ready. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) ps || true; exit 1)

	@if [ -n "$(WEB_LOGIN_HEALTH_URL)" ]; then \
		if command -v gum >/dev/null 2>&1; then \
			gum spin --spinner dot --title "waiting for web-login $(WEB_LOGIN_HEALTH_URL)..." -- \
				bash -lc '\
					deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
					while [ $$SECONDS -lt $$deadline ]; do \
						if command -v curl >/dev/null 2>&1 && curl -fsS "$(WEB_LOGIN_HEALTH_URL)" >/dev/null; then exit 0; fi; \
						sleep 0.5; \
					done; \
					exit 1'; \
		else \
			deadline=$$((SECONDS+$(WAIT_TIMEOUT_SEC))); \
			while [ $$SECONDS -lt $$deadline ]; do \
				if command -v curl >/dev/null 2>&1 && curl -fsS "$(WEB_LOGIN_HEALTH_URL)" >/dev/null; then break; fi; \
				sleep 0.5; \
			done; \
		fi; \
	fi \
	|| (echo "Web-login not ready. Current service states:"; $(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) ps || true; exit 1)

	@echo "✓ Web started"
	@$(COMPOSE) $(COMPOSE_ARGS_WITH_WEB) ps

platform-image-build:
	@docker build -t colabsphere-api:$(IMAGE_TAG) -f platform/Dockerfile platform

web-image-build:
	@docker build -t collabsphere-web:$(IMAGE_TAG) -f web/Dockerfile web

migrate:
	@$(MAKE) platform-image-build
	@mkdir -p $(dir $(MIGRATE_LOG))
	@docker compose $(MIGRATE_COMPOSE_ARGS) \
		run --rm --build -e MIGRATE_CMD=$(MIGRATE_CMD) migrate \
		> $(MIGRATE_LOG) 2>&1 \
	|| (echo "migrate failed; tail:"; tail -n 160 $(MIGRATE_LOG) || true; exit 1)

seed:
	@$(MAKE) platform-image-build
	@mkdir -p $(dir $(SEED_LOG))
	@docker compose $(SEED_COMPOSE_ARGS) \
		run --rm --build -e SEED_CMD=$(SEED_CMD) seed \
		> $(SEED_LOG) 2>&1 \
	|| (echo "seed failed; tail:"; tail -n 160 $(SEED_LOG) || true; exit 1)

migrations-build:
	go -C platform run ./internal/runtime/infrastructure/db/cmd/build-migrations

check-migrations:
	./scripts/check-migrations.sh

check-contracts:
	@mkdir -p /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts check-parity

contracts-openapi-json:
	@mkdir -p /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts openapi-json

contracts-openapi-yaml:
	@mkdir -p /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts openapi-yaml

contracts-routes:
	@mkdir -p /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts routes

contracts-snapshot:
	@mkdir -p docs/openapi /tmp/collabsphere-go-cache /tmp/collabsphere-go-modcache
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts openapi-json > docs/openapi/openapi.json
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts openapi-yaml > docs/openapi/openapi.yaml
	@$(CONTRACTS_ENV) GOCACHE=/tmp/collabsphere-go-cache GOMODCACHE=/tmp/collabsphere-go-modcache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform run ./cmd/contracts routes > docs/openapi/routes.txt

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

test-accounts-integration:
	@if [ -z "$$COLLABSPHERE_TEST_POSTGRES_DSN" ]; then \
		echo "COLLABSPHERE_TEST_POSTGRES_DSN is required"; \
		exit 1; \
	fi
	@mkdir -p /tmp/collabsphere-go-cache
	@GOCACHE=/tmp/collabsphere-go-cache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform test -tags=integration ./internal/accounts/delivery/http -run 'TestCreateAccountIntegration'

test-organizations-integration:
	@if [ -z "$$COLLABSPHERE_TEST_POSTGRES_DSN" ]; then \
		echo "COLLABSPHERE_TEST_POSTGRES_DSN is required"; \
		exit 1; \
	fi
	@mkdir -p /tmp/collabsphere-go-cache
	@GOCACHE=/tmp/collabsphere-go-cache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform test -tags=integration ./internal/organizations/delivery/http -run 'Test.*Organization.*Integration'

test-platform-reviews-integration:
	@if [ -z "$$COLLABSPHERE_TEST_POSTGRES_DSN" ]; then \
		echo "COLLABSPHERE_TEST_POSTGRES_DSN is required"; \
		exit 1; \
	fi
	@mkdir -p /tmp/collabsphere-go-cache
	@GOCACHE=/tmp/collabsphere-go-cache TMPDIR=/tmp \
		/usr/local/go/bin/go -C platform test -tags=integration ./internal/platformops/delivery/http -run 'Test.*Review.*Integration'

smoke-account-signup-login: up-dev migrate-up
	@account_email="smoke+$$(date +%s)@example.com"; \
	account_password="Secret123"; \
	echo "Smoke signup/login for $$account_email"; \
	curl -fsS \
		-H 'Content-Type: application/json' \
		-d "{\"email\":\"$$account_email\",\"password\":\"$$account_password\"}" \
		http://localhost:8080/v1/accounts >/dev/null; \
	curl -fsS \
		-H 'Content-Type: application/json' \
		-d "{\"email\":\"$$account_email\",\"password\":\"$$account_password\"}" \
		http://localhost:8080/v1/auth/login >/dev/null; \
	echo "✓ Smoke signup/login passed"

smoke-bootstrap:
	@./scripts/smoke-bootstrap.sh

smoke-auth-legacy:
	@./scripts/smoke-auth-legacy.sh

smoke-auth-zitadel-e2e:
	@./scripts/smoke-auth-zitadel-e2e.sh

baseline-metrics:
	@./scripts/baseline-metrics.sh
