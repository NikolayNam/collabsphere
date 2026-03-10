# CollabSphere

CollabSphere - backend-платформа на Go для управления аккаунтами, организациями и участниками. Репозиторий содержит HTTP API, миграции PostgreSQL, Docker Compose-конфигурацию для локального запуска и архитектурные документы.

Этот `README.md` описывает текущее состояние проекта. Он не пытается скрыть ограничения: часть инфраструктуры уже собрана, HTTP API для базовых модулей доступно, а auth-контур пока остается незавершенным.

## Что есть в репозитории

- Go 1.26 API с роутингом через `chi` и OpenAPI-документацией через `huma`
- PostgreSQL + `gorm` + `pgx`
- Миграции через `goose`
- Docker Compose-окружение в каталоге `deploy/`
- Бизнес-модули `accounts`, `organizations`, `memberships`, `auth`
- Архитектурные заметки и ADR в `docs/architecture/`

## Стек и модули

### Технологии

| Область | Текущая технология |
| --- | --- |
| HTTP API | Go 1.26, `chi`, `huma` |
| OpenAPI / docs | `huma`, Scalar API reference на `/api/v1/docs` |
| Доступ к БД | `gorm`, `pgx` |
| База данных | PostgreSQL |
| Миграции | `goose` + генерация bundled SQL |
| Локальный запуск | Docker Compose |

### HTTP-модули

- `accounts`: создание аккаунта, поиск по ID и email
- `organizations`: создание организации и получение по ID
- `memberships`: добавление и просмотр участников организации
- `auth`: маршруты зарегистрированы, но текущая реализация не считается завершенной

### Доменные области в БД

Помимо активного HTTP-слоя, в миграциях уже присутствуют заготовки для `catalog`, `sales`, `storage` и расширенного `auth`. Это важно учитывать: база шире, чем текущая публичная API-поверхность.

## Структура репозитория и архитектурные правила

```text
collabsphere/
  deploy/      Docker Compose, env-файл, secrets
  docs/        ADR и служебная документация
  platform/    Go-код приложения
```

Ключевые части Go-кода:

- `platform/cmd/api` - entrypoint HTTP API
- `platform/cmd/migrate` - entrypoint мигратора
- `platform/internal/runtime/bootstrap` - composition root и сборка приложения
- `platform/internal/runtime/foundation` - базовые примитивы runtime-слоя
- `platform/internal/runtime/infrastructure` - HTTP server, middleware, DB plumbing, transport adapters
- `platform/internal/accounts`, `organizations`, `memberships`, `auth` - бизнес-модули
- `platform/shared` - переиспользуемые библиотеки без зависимости на `internal/*`

Архитектурные границы и правила размещения кода зафиксированы в ADR:
[`docs/architecture/adr-foundation-infrastructure-boundaries.md`](docs/architecture/adr-foundation-infrastructure-boundaries.md)

## Требования

### Обязательные

- Docker Engine / Docker Desktop с Compose v2
- Go `1.26.x`
- доступ к `docker compose`

### Для Linux / WSL сценария через Makefile

- `make`
- `bash`
- желательно `curl` для health-check в `make up-dev`

### Опционально

- `gum` для spinner в `make up-dev`
- `codeweaver` для генерации `docs/codebase_actual.md`

## Быстрый старт

Все команды ниже выполняются из корня репозитория.

### Linux / WSL

Рекомендуемый путь для локальной разработки:

```bash
docker network create external.network >/dev/null 2>&1 || true
make up-dev
make migrate-up
```

После запуска:

- Scalar API reference: [http://localhost:8080/api/v1/docs](http://localhost:8080/api/v1/docs)
- OpenAPI YAML: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
- Health-check: [http://localhost:8080/health](http://localhost:8080/health)

Полезные команды:

```bash
go -C platform test ./...
make migrate-down
```

Логи `make up-dev` и `make migrate-up` пишутся в `logs/docker/`.

### Windows PowerShell

`Makefile` ориентирован на `bash`, поэтому в Windows проще работать через прямые `docker compose` команды.

```powershell
docker network create external.network 2>$null

docker compose `
  --env-file deploy/.env.dev `
  -f deploy/docker-compose.postgres.yaml `
  -f deploy/docker-compose.platform.yaml `
  --profile local up -d --build --force-recreate

# если нужен локальный ZITADEL для OIDC-login
# сначала заполните AUTH_ZITADEL_* и ZITADEL_* в deploy/.env.dev
# затем поднимите дополнительный compose-файл

docker compose `
  --env-file deploy/.env.dev `
  -f deploy/docker-compose.postgres.yaml `
  -f deploy/docker-compose.platform.yaml `
  -f deploy/docker-compose.zitadel.yaml `
  --profile local up -d --build --force-recreate

$env:MIGRATE_CMD = "up"
docker compose `
  --env-file deploy/.env.dev `
  -p collabsphere-migrate `
  -f deploy/docker-compose.migrate.yaml `
  --profile migrate run --rm --build migrate
```

Проверка после запуска:

- [http://localhost:8080/api/v1/docs](http://localhost:8080/api/v1/docs)
- [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
- [http://localhost:8080/health](http://localhost:8080/health)

## Локальный ZITADEL

Для dev-окружения добавлен отдельный compose-файл:

- `deploy/docker-compose.zitadel.yaml`

Он поднимает:

- `zitadel-db` - отдельный PostgreSQL только для ZITADEL
- `zitadel` - сам IAM server
- `zitadel-login` - login UI

Минимальная схема локального запуска такая:

1. Заполнить блок `AUTH_ZITADEL_*` и `ZITADEL_*` в `deploy/.env.dev` (`ZITADEL_MASTERKEY` должен быть ровно 32 ASCII-символа)
2. Поднять стек вместе с `deploy/docker-compose.zitadel.yaml`
3. При необходимости заранее переопределить bootstrap-поля первого администратора: `ZITADEL_FIRSTINSTANCE_ORG_HUMAN_USERNAME`, `ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_ADDRESS`, `ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_VERIFIED`, `ZITADEL_FIRSTINSTANCE_ORG_HUMAN_PASSWORD`
4. Если ZITADEL уже запускался с неверным hostname или старыми настройками Login V2, удалить локальные тома `zitadel.postgres.data` и `zitadel.shared`, затем повторить первый старт
5. Создать в ZITADEL OIDC application для backend callback `http://localhost:8080/api/v1/auth/zitadel/callback`
6. Перенести выданные `client_id` и `client_secret` в `AUTH_ZITADEL_CLIENT_ID` и `AUTH_ZITADEL_CLIENT_SECRET`, затем установить `AUTH_ZITADEL_ENABLED=true`
7. Перезапустить `api`

Для `ZITADEL_MASTERKEY` используйте случайный ASCII-ключ ровно на 32 символа. В PowerShell его можно сгенерировать так:

```powershell
-join ((48..57 + 65..90 + 97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

`AUTH_ZITADEL_CLIENT_ID` и `AUTH_ZITADEL_CLIENT_SECRET` не генерируются заранее в `.env.dev`. Они появляются только после первого запуска ZITADEL, когда вы создаёте OIDC application в админ-панели и копируете оттуда выданные значения.


По умолчанию используется hostname `auth.localhost`.
Это сделано намеренно: браузер на хосте должен открывать `http://auth.localhost:8090` и `http://auth.localhost:3000`, а контейнер `api` получает тот же hostname через `extra_hosts` в `deploy/docker-compose.platform.yaml`. Использование `localhost:3000` для Login V2 приводит к `Instance not found`, потому что инстанс ZITADEL зарегистрирован на `auth.localhost`.
`ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_ADDRESS` и `ZITADEL_FIRSTINSTANCE_ORG_HUMAN_EMAIL_VERIFIED` применяются только на первом bootstrap инстанса. Если ZITADEL уже инициализирован, изменение `.env.dev` само по себе не обновит существующего администратора.
### Остановка окружения

```bash
docker compose \
  -f deploy/docker-compose.postgres.yaml \\
  -f deploy/docker-compose.platform.yaml \
  --profile local down
```

## Конфигурация и секреты

### Источники конфигурации

Текущая Go-конфигурация читается из environment variables и на данный момент использует только:

- `TZ`
- `APPLICATION_TITLE`
- `APPLICATION_VERSION`
- `APPLICATION_ADDRESS` (по умолчанию `0.0.0.0:8080`)
- `APPLICATION_TIMEOUT_READ`
- `APPLICATION_TIMEOUT_WRITE`
- `APPLICATION_TIMEOUT_IDLE`
- `APPLICATION_DEBUG`
- `POSTGRES_HOST`
- `POSTGRES_PORT`
- `POSTGRES_DB`
- `POSTGRES_SCHEMA`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD` или `POSTGRES_PASSWORD_FILE`
- `POSTGRES_DEBUG`

`deploy/.env` используется Compose-окружением и содержит больше переменных, чем реально читает текущий Go runtime. Считайте этот файл инфраструктурным шаблоном, а не полной спецификацией приложения.

### Важное различие между `APPLICATION_ADDRESS` и `APPLICATION_PORT`

- `APPLICATION_ADDRESS` - это фактический адрес, на котором слушает Go HTTP server
- `APPLICATION_PORT` в `deploy/.env` сейчас нужен Compose-конфигурации для проброса порта `${APPLICATION_PORT}:${APPLICATION_PORT}`

Если `APPLICATION_ADDRESS` не задан, API по коду слушает `0.0.0.0:8080`.

### Secrets, на которые ссылается Compose

Compose-файлы ожидают локальные файлы в `deploy/secrets/`, в том числе:

- `deploy/secrets/postgres/dev/db_password`
- `deploy/secrets/postgres/prod/db_password`
- `deploy/secrets/jwt/auth_jwt_secret`

Даже для локального профиля часть секретов объявлена на уровне сервиса, поэтому отсутствие placeholder-файлов может ломать запуск Compose.

## Миграции

### Где лежат миграции

- исходники миграций: `platform/internal/runtime/infrastructure/db/migrations-src/`
- сгенерированный bundle: `platform/internal/runtime/infrastructure/db/migrations/`
- порядок сборки задается через `platform/internal/runtime/infrastructure/db/migrations-src/manifest.yaml`

### Базовые команды

```bash
make migrations-build
make migrate-up
make migrate-down
```

`make migrate-up` и `make migrate-down` используют `deploy/docker-compose.migrate.yaml` и запускают контейнер с `platform/cmd/migrate`.

Если вы меняете SQL в `migrations-src/`, сначала пересоберите bundle через `make migrations-build`, а уже потом прогоняйте миграции.

### Важно для `goose`

Если в SQL-миграции используется `DO $$ ... $$;`, оборачивайте блок:

```sql
-- +goose StatementBegin
DO $$
BEGIN
  -- SQL / PLpgSQL
END
$$;
-- +goose StatementEnd
```

Иначе `goose` может разрезать выражение по `;` и завершиться ошибкой парсинга.

## Тесты

Минимальная быстрая проверка приложения:

```bash
go -C platform test ./...
```

Это же полезно запускать перед сборкой Docker-образа. В `platform/Dockerfile` дополнительно выполняются `go vet ./...` и `go test -v ./...` на build stage.

## API и документация

### Базовый префикс

Все зарегистрированные маршруты API живут под префиксом `/api/v1`.

### Redirects из корня

Корневой router дополнительно пробрасывает удобные entrypoints:

- `/openapi.yaml` -> `/api/v1/openapi.yaml`
- `/health` -> `/api/v1/health`

### Основные группы маршрутов

Системные:

- `GET /api/v1/health`

Accounts:

- `POST /api/v1/accounts`
- `GET /api/v1/accounts/{id}`
- `GET /api/v1/accounts/by-email`

Organizations:

- `POST /api/v1/organizations`
- `GET /api/v1/organizations/{id}`

Memberships:

- `POST /api/v1/organizations/{organization_id}/members`
- `GET /api/v1/organizations/{organization_id}/members`

Auth:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

Маршруты `auth` есть в OpenAPI и регистрируются в приложении, но текущую реализацию нельзя считать стабилизированной. Подробности см. в разделе ниже.

## Known issues

- Старый корневой README был неполным и частично расходился с текущим кодом и Compose-конфигурацией. Этот документ заменяет его целиком.
- `make network` сейчас создает сеть `platform.web.network`, тогда как Compose-файлы ожидают внешнюю сеть `web.network`. Для безопасного запуска создавайте сеть вручную: `docker network create web.network`.
- `Makefile` ориентирован на Linux / WSL и использует `bash`. В PowerShell и CMD лучше вызывать `docker compose` напрямую.
- Auth-контур пока не завершен: маршруты зарегистрированы, но в bootstrap токен-менеджер не подключен, поэтому auth-flow не стоит считать production-ready.
- В `deploy/.env` есть расхождения с текущим runtime-кодом, включая строку `AUTH_JWT_SECRET_FILE==...`. Кроме того, часть переменных из файла сейчас не читается приложением вообще.
- Compose-конфигурация и env уже содержат задел под будущие подсистемы и внешние интеграции. Не вся эта конфигурация соответствует текущему фактическому поведению API.

## Полезные ссылки

- ADR по архитектурным границам: [`docs/architecture/adr-foundation-infrastructure-boundaries.md`](docs/architecture/adr-foundation-infrastructure-boundaries.md)
- Compose-файлы: `deploy/docker-compose.postgres.yaml`, `deploy/docker-compose.platform.yaml`, `deploy/docker-compose.migrate.yaml`
- Исходники API: `platform/cmd/api`, `platform/internal/`











