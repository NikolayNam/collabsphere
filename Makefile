COMPOSE := docker compose
EXEC := docker exec -it
LOGS := docker logs
ENV_FILE := --env-file .env
MIGRATE_CMD ?= up

PROJECT_NAME := collabsphere
DEPLOY_DIR := deploy


INFRA_FILE := docker-compose.infrastructure.yaml
PLATFORM_FILE := docker-compose.platform.yaml
MIGRATE_FILE := docker-compose.migrate.yaml

NETWORK_NAME := platform.web.network

LOG_DIR := ./logs/docker
APP_LOG := $(LOG_DIR)/app.log
MIGRATE_LOG := $(LOG_DIR)/migrate.log
CODEBASE_LOG := $(LOG_DIR)/codebase.log

CODEBASE_OUTPUT := ./docs/codebase_actual.md

COMPOSE_ARGS = \
	-f $(DEPLOY_DIR)/$(INFRA_FILE) \
	-f $(DEPLOY_DIR)/$(PLATFORM_FILE) \
	--profile local

SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

API_HEALTH_URL ?= http://localhost:8080/health
WAIT_TIMEOUT_SEC ?= 60

MIGRATE_COMPOSE_ARGS = \
	-p collabsphere-migrate \
	-f deploy/docker-compose.migrate.yaml \
	--profile migrate

.PHONY: cloudsphere-init network up-app up-dev up-prod down logs sync migrate clean-logs

cloudsphere-init: network up-dev migrate

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
	@MIGRATE_CMD=$(MIGRATE_CMD) docker compose $(MIGRATE_COMPOSE_ARGS) \
		run --rm --build migrate \
		> $(MIGRATE_LOG) 2>&1 \
	|| (echo "migrate failed; tail:"; tail -n 160 $(MIGRATE_LOG) || true; exit 1)


migrate-up:
	@$(MAKE) migrate MIGRATE_CMD=up

migrate-down:
	@$(MAKE) migrate MIGRATE_CMD=down

codebase:
	@mkdir -p $(LOG_DIR)
	@mkdir -p ./docs
	@echo Generating codebase markdown...
	@rm -f $(CODEBASE_OUTPUT)
	@codeweaver -input=. -output=$(CODEBASE_OUTPUT) \
		-include="\\.go$$,\\.md$$,\\.sql$$,\\.yaml$$" \
		-ignore="^\\.git,^docs/" \
		> $(CODEBASE_LOG) 2>&1


logs:
	$(LOGS) -f $(APP_CONTAINER)