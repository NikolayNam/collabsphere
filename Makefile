



# Makefile (CollabSphere)

DC            ?= docker compose
ENV_FILE      ?= .env

INFRA_FILE    ?= docker-compose.infrastructure.yaml
PLATFORM_FILE ?= docker-compose.platform.yaml
APP_FILE      ?= docker-compose/docker-compose.yaml

NETWORK       ?= platform.web.network

APP_CONTAINER ?= cloudsphere-backend

PROFILE_DEV   ?= api.local
PROFILE_PROD  ?= api.cloud

.PHONY: help
help:
	@echo "Targets:"
	@echo "  init            - create network (idempotent)"
	@echo "  up              - up using APP_FILE (legacy single compose)"
	@echo "  up-dev          - up infra+platform with dev profile"
	@echo "  up-prod         - up infra+platform with prod profile"
	@echo "  down            - down infra+platform (all profiles)"
	@echo "  down-app        - down APP_FILE (legacy single compose)"
	@echo "  logs            - follow logs for APP_CONTAINER"
	@echo "  ps              - list running containers"
	@echo "  exec            - exec into APP_CONTAINER (sh)"
	@echo ""
	@echo "Vars:"
	@echo "  ENV_FILE=.env APP_CONTAINER=... PROFILE_DEV=... PROFILE_PROD=..."

.PHONY: init
init:
	@$(DC) network create $(NETWORK) >/dev/null 2>&1 || true

# --- Legacy single-compose flow (if you still need it) ---

.PHONY: up
up: init
	$(DC) -f $(APP_FILE) --env-file $(ENV_FILE) up -d --build

.PHONY: down-app
down-app:
	$(DC) -f $(APP_FILE) --env-file $(ENV_FILE) down

# --- Main flow: infra + platform with profiles ---

define COMPOSE_STACK
$(DC) -f $(INFRA_FILE) -f $(PLATFORM_FILE) --env-file $(ENV_FILE)
endef

.PHONY: up-dev
up-dev: init
	$(COMPOSE_STACK) --profile $(PROFILE_DEV) up -d --build --force-recreate

.PHONY: up-prod
up-prod: init
	$(COMPOSE_STACK) --profile $(PROFILE_PROD) up -d --build --force-recreate

.PHONY: down
down:
	$(COMPOSE_STACK) down

.PHONY: ps
ps:
	$(DC) ps

.PHONY: logs
logs:
	docker logs -f $(APP_CONTAINER)

.PHONY: exec
exec:
	docker exec -it $(APP_CONTAINER) sh