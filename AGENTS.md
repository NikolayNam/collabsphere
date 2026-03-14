# CollabSphere Repository Guidelines

Этот репозиторий нужно эволюционировать, а не переписывать.

## Что сохраняем

- layout репозитория: `deploy/`, `docs/`, `platform/`
- основной Go-код: `platform/cmd`, `platform/internal/runtime`, `platform/internal/<domain>`, `platform/shared`
- стек: Go `1.26`, `chi`, `huma`, PostgreSQL, `gorm`, `pgx`, `goose`, Docker Compose
- API-префикс: `/v1`
- pipeline миграций: `migrations-src/` + bundled migrations
- архитектурные границы из `docs/architecture/` и ADR

## Приоритеты этого цикла

1. Доводить и стабилизировать текущий backend core.
2. Сначала auth, bootstrap, config, docs, tests.
3. Затем accounts, organizations, memberships и уже существующие control-plane маршруты.
4. Только после стабилизации ядра расширять platform-функциональность.

## Что не делать без веской причины

- не переносить проект на другую корневую архитектуру
- не вводить microservices
- не заменять стековые компоненты без жёсткого технического обоснования
- не начинать broad universal-core modules, пока auth и core flows не стабилизированы

## Правила по изменениям

- работайте внутри существующих модулей и bootstrap wiring
- не ломайте текущие `/v1` health/OpenAPI routes
- новые миграции делайте через текущий pipeline и не забывайте bundled migrations
- README, docs и OpenAPI должны соответствовать реальному поведению API
- изменения должны сопровождаться честными заметками об ограничениях, если они остались
- `docs/post-stabilization-dod.md`, `docs/configuration.md` и `docs/openapi/*` считаются частью стабилизационного контракта и должны обновляться вместе с поведением API

## Тесты и проверка

- обязательно прогоняйте релевантные unit/integration-тесты для затронутых flow
- критические integration-сценарии: `auth`, `accounts`, `organizations`, `memberships`, существующие `platformops`
- для контрактов используйте `go -C platform run ./cmd/contracts ...` или `make check-contracts`
- если локальный `go` резолвится неверно, в этой среде используйте `/usr/local/go/bin/go`

## Осторожность с окружением

- compose/env/secrets чувствительны; не меняйте `deploy/compose/*.yaml`, `deploy/.env*` и secret wiring без явного запроса пользователя
- не выдавайте неподтверждённые claims о production-ready статусе, если flow зависит от внешнего окружения или секретов
