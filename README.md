# CollabSphere

CollabSphere - backend-first платформа на Go для управления аккаунтами, организациями и участниками. Репозиторий содержит HTTP API, миграции PostgreSQL, Docker Compose-конфигурацию для локального запуска, архитектурные документы и новый `web/` frontend shell на `Next.js`.

Этот `README.md` описывает текущее состояние проекта. Он не пытается скрывать ограничения: базовые backend-модули и auth-контур собраны и покрыты критическими тестами, но полная production-готовность всё ещё зависит от локальной конфигурации окружения, секретов и внешних интеграций вроде ZITADEL.

## Что есть в репозитории

- Go 1.26 API с роутингом через `chi` и OpenAPI-документацией через `huma`
- `Next.js` frontend shell в `web/`, работающий поверх текущего Go API
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
| Frontend shell | `Next.js` App Router, TypeScript |
| OpenAPI / docs | `huma`, Scalar API reference на `/v1/docs` |
| Доступ к БД | `gorm`, `pgx` |
| База данных | PostgreSQL |
| Миграции | `goose` + генерация bundled SQL |
| Локальный запуск | Docker Compose |

### HTTP-модули

- `accounts`: создание аккаунта, поиск по ID и email
- `organizations`: создание организации и получение по ID
- `memberships`: добавление и просмотр участников организации, invitations и acceptance flow
- `auth`: browser login через ZITADEL, exchange, refresh rotation с reuse detection, logout, `me` и legacy email/password fallback под feature flags

### Доменные области в БД

Помимо активного HTTP-слоя, в миграциях уже присутствуют заготовки для `catalog`, `sales`, `storage` и расширенного `auth`. Это важно учитывать: база шире, чем текущая публичная API-поверхность.

## Структура репозитория и архитектурные правила

```text
collabsphere/
  deploy/      Docker Compose entrypoints, env/, observability/, secrets/
  docs/        ADR и служебная документация
  platform/    Go-код приложения
  web/         Next.js frontend shell
```

Ключевые части Go-кода:

- `platform/cmd/api` - entrypoint HTTP API
- `platform/cmd/migrate` - entrypoint мигратора
- `platform/internal/runtime/bootstrap` - composition root и сборка приложения
- `platform/internal/runtime/foundation` - базовые примитивы runtime-слоя
- `platform/internal/runtime/infrastructure` - HTTP server, middleware, DB plumbing, transport adapters
- `platform/internal/accounts`, `organizations`, `memberships`, `auth` - бизнес-модули
- `platform/shared` - переиспользуемые библиотеки без зависимости на `internal/*`
- `web/app` - frontend routes и UI над существующим `/v1` API
- `web/lib` - frontend helpers для auth/API, не заменяющие backend-логику
- отдельный frontend README: [`web/README.md`](web/README.md)

Архитектурные границы и правила размещения кода зафиксированы в ADR:
[`docs/architecture/adr-foundation-infrastructure-boundaries.md`](docs/architecture/adr-foundation-infrastructure-boundaries.md)

Стабилизационная планка этого цикла вынесена отдельно:

- [`docs/post-stabilization-dod.md`](docs/post-stabilization-dod.md)
- [`docs/configuration.md`](docs/configuration.md)

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
- `Node.js >= 20.9` и `npm` для `web/`

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

- Scalar API reference: [http://localhost:8080/v1/docs](http://localhost:8080/v1/docs)
- OpenAPI YAML: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
- Health-check: [http://localhost:8080/health](http://localhost:8080/health)
- Readiness: [http://localhost:8080/v1/ready](http://localhost:8080/v1/ready)

Полезные команды:

```bash
go -C platform test ./...
make check-contracts
make contracts-snapshot
make smoke-bootstrap
make smoke-auth-legacy
make smoke-auth-zitadel-e2e
make migrate-down
```

Логи `make up-dev` и `make migrate-up` пишутся в `logs/docker/`.

### Разделение env-файлов

Локальный dev-стек теперь использует overlay-структуру в `deploy/env/`:

- `deploy/env/.env.dev` — базовый backend/platform слой
- `deploy/env/.env.postgres.dev` — PostgreSQL overlay
- `deploy/env/.env.storage.dev` — storage и S3
- `deploy/env/.env.redis.dev` — realtime/redis
- `deploy/env/.env.zitadel.dev` — ZITADEL/OIDC
- `deploy/env/.env.web.dev` — frontend `web/`

Это сделано специально, чтобы `postgres`, `storage`, `zitadel`, `redis` и `web` не смешивались в одном большом `.env.dev`.

### Frontend `web/`

Новый frontend лежит в `web/` и не заменяет backend. Он использует уже существующий auth flow:

1. browser login уходит в backend `GET /v1/auth/zitadel/login`
2. backend после OIDC callback возвращает пользователя на frontend `auth/callback`
3. frontend вызывает `POST /v1/auth/exchange`
4. дальше работает уже с локальными backend токенами

Локальный запуск:

```bash
cd web
cp .env.local.example .env.local
npm install
npm run dev
```

Или поднять frontend контейнером вместе со всей платформой:

```bash
make up-dev-web
```

Этот сценарий использует overlay-набор env-файлов:

- `deploy/env/.env.dev` для базового backend/platform стека
- `deploy/env/.env.postgres.dev` для PostgreSQL
- `deploy/env/.env.storage.dev` для storage/S3
- `deploy/env/.env.redis.dev` для realtime/redis
- `deploy/env/.env.zitadel.dev` для ZITADEL/OIDC
- `deploy/env/.env.web.dev` для frontend-специфичных `WEB_*`

Ожидаемые адреса:

- frontend: [http://collabsphere.localhost:3002](http://collabsphere.localhost:3002)
- backend API: [http://api.localhost:8080](http://api.localhost:8080)

Что важно:

- frontend XHR к backend идут через `Next.js rewrites` на `/api/backend/*`
- это сделано намеренно, потому что в текущем backend runtime нет отдельного выделенного CORS-контура
- browser navigation в `GET /v1/auth/zitadel/login` идёт напрямую в backend

Текущий MVP-срез `web/` включает:

- `/`
- `/login`
- `/auth/callback`
- `/me`
- `/organizations`
- `/chat`

Сейчас frontend хранит токены в `localStorage` как быстрый MVP-компромисс. Для production-ready контура это нужно будет заменить на `httpOnly` cookie/BFF слой.

Для containerized запуска frontend включается через профиль `web` в:

- `deploy/compose/core.yaml`

Этот профиль не только поднимает `web`/`web-login`, но и переопределяет `AUTH_BROWSER_REDIRECT_ORIGINS` у `api` под frontend origin, чтобы browser login не упирался в redirect allowlist.

Полезные overrides:

- `WEB_HOST_PORT` — host port frontend, по умолчанию `3002`
- `WEB_NEXT_PUBLIC_APP_BASE_URL` — внешний frontend origin, по умолчанию `http://collabsphere.localhost:3002`
- `WEB_NEXT_PUBLIC_API_BASE_URL` — внешний backend origin для browser login/callback, по умолчанию `http://api.localhost:8080`
- `WEB_NEXT_INTERNAL_API_BASE_URL` — backend origin внутри Docker network, по умолчанию `http://api:8080`
- `WEB_AUTH_BROWSER_REDIRECT_ORIGINS` — allowlist origins для backend browser auth при включённом `deploy/compose/core.yaml` c профилем `web`; по умолчанию `http://api.localhost:8080,http://collabsphere.localhost:3002`

### Prometheus

Prometheus здесь используется для метрик, а не для хранения сырых request-логов. Для HTTP API он собирает:

- `collabsphere_http_requests_total`
- `collabsphere_http_request_duration_seconds`
- `collabsphere_http_response_size_bytes`
- `collabsphere_http_requests_in_flight`

Чтобы включить endpoint метрик в API, добавьте в `deploy/env/.env.dev`:

```env
APPLICATION_METRICS_ENABLED=true
APPLICATION_METRICS_PATH=/metrics
```

После этого поднимите API как обычно и отдельно запустите Prometheus:

```bash
docker compose --env-file deploy/env/.env.dev -f deploy/compose/observability.yaml --profile observability up -d prometheus
```

Проверка:

- API metrics: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- Prometheus UI: [http://localhost:9000](http://localhost:9000)

По умолчанию шумные инфраструктурные запросы на `/health`, `/v1/health` и `/metrics` не пишутся в access-log приложения, чтобы не забивать логи healthcheck'ами и scrape'ами.

### Grafana + Loki + Alloy

Для централизованного сбора логов API добавлен отдельный observability-стек:

- `deploy/compose/observability.yaml`
- `deploy/observability/loki/loki-config.yaml`
- `deploy/observability/alloy/config.alloy`
- `deploy/observability/grafana/provisioning/...`

Первый rollout ограничен только контейнером `api`. `worker` и инфраструктурные контейнеры в Loki пока не отправляются.

Что важно:

- API по-прежнему пишет структурированные JSON-логи в `stdout`
- Alloy читает Docker logs только у `collabsphere/api`
- Loki хранит логи
- Grafana подключается и к Loki, и к уже существующему Prometheus
- request bodies не логируются
- чувствительные поля вроде `Authorization`, `Cookie`, `password`, `refresh_token`, `client_secret` в этот MVP не собираются вообще

Запуск:

```bash
# API должен уже работать, а Prometheus - быть поднят отдельно
docker compose --env-file deploy/env/.env.dev -f deploy/compose/observability.yaml --profile observability up -d
```

Проверка:

- Grafana: [http://localhost:3001](http://localhost:3001)
- Loki datasource провиженится автоматически
- Prometheus datasource смотрит на `http://host.docker.internal:9000`

В Grafana уже провиженится стартовый dashboard `CollabSphere API Logs` с панелями:

- HTTP `4xx/5xx` по статусам
- recent HTTP errors
- auth failures
- DB/request errors
- all API logs

Для поиска по конкретному `request_id` используйте Explore в Grafana с Loki-запросом:

```logql
{service="api"} | json | request_id="<request-id>"
```

Для работы этой связки должны быть включены API-метрики:

```env
APPLICATION_METRICS_ENABLED=true
APPLICATION_METRICS_PATH=/metrics
```

### Windows PowerShell

`Makefile` ориентирован на `bash`, поэтому в Windows проще работать через прямые `docker compose` команды.

```powershell
docker network create external.network 2>$null

docker compose `
  --env-file deploy/env/.env.dev `
  --env-file deploy/env/.env.postgres.dev `
  --env-file deploy/env/.env.storage.dev `
  --env-file deploy/env/.env.redis.dev `
  --env-file deploy/env/.env.zitadel.dev `
  -f deploy/compose/core.yaml `
  -f deploy/compose/storage.yaml `
  -f deploy/compose/auth.yaml `
  --profile local up -d --build --force-recreate

# если нужен frontend в контейнере вместе с платформой
docker compose `
  --env-file deploy/env/.env.dev `
  --env-file deploy/env/.env.postgres.dev `
  --env-file deploy/env/.env.storage.dev `
  --env-file deploy/env/.env.redis.dev `
  --env-file deploy/env/.env.zitadel.dev `
  --env-file deploy/env/.env.web.dev `
  -f deploy/compose/core.yaml `
  -f deploy/compose/storage.yaml `
  -f deploy/compose/auth.yaml `
  --profile local --profile web up -d --build --force-recreate

# если нужен локальный ZITADEL для OIDC-login
# сначала заполните AUTH_ZITADEL_* и ZITADEL_* в deploy/env/.env.zitadel.dev
# затем поднимите дополнительный compose-файл

docker compose `
  --env-file deploy/env/.env.dev `
  --env-file deploy/env/.env.postgres.dev `
  --env-file deploy/env/.env.storage.dev `
  --env-file deploy/env/.env.redis.dev `
  --env-file deploy/env/.env.zitadel.dev `
  -f deploy/compose/core.yaml `
  -f deploy/compose/storage.yaml `
  -f deploy/compose/auth.yaml `
  --profile local up -d --build --force-recreate

$env:MIGRATE_CMD = "up"
docker compose `
  --env-file deploy/env/.env.dev `
  --env-file deploy/env/.env.postgres.dev `
  -p collabsphere-migrate `
  -f deploy/compose/jobs.yaml `
  --profile migrate run --rm --build migrate
```

Проверка после запуска:

- [http://localhost:8080/v1/docs](http://localhost:8080/v1/docs)
- [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
- [http://localhost:8080/health](http://localhost:8080/health)

## Локальный ZITADEL

Для dev-окружения identity-стек живёт в отдельном compose-файле:

- `deploy/compose/auth.yaml`

Он поднимает:

- `zitadel-db` - отдельный PostgreSQL только для ZITADEL
- `zitadel` - сам IAM server
- `web-login` - self-hosted login UI on the `auth.localhost` origin

Минимальная схема локального запуска такая:

1. Заполнить публичные `AUTH_ZITADEL_*` и `ZITADEL_*` в `deploy/env/.env.zitadel.dev`
2. Заполнить локальные secret-файлы в `deploy/secrets/identity/`: `zitadel_master_key`, `zitadel_runtime_secrets.yaml`, `zitadel_init_steps.yaml`
3. Поднять стек вместе с `deploy/compose/auth.yaml`
4. При необходимости заранее переопределить bootstrap-поля первого администратора в `deploy/secrets/identity/zitadel_init_steps.yaml`
5. Если ZITADEL уже запускался с неверным hostname или старыми настройками Login V2, удалить локальные тома `zitadel.postgres.data` и `zitadel.shared`, затем повторить первый старт
   Для self-hosted login это часто проявляется как `500` в `web-login` или proxy/login action responses с ошибками ZITADEL вроде `Errors.Token.Invalid (AUTH-7fs1e)` при вызове session APIs
6. Создать в ZITADEL OIDC application для backend callback `http://api.localhost:8080/v1/auth/zitadel/callback`
7. Перенести выданные `client_id` в `AUTH_ZITADEL_CLIENT_ID`, а `client_secret` в `deploy/secrets/identity/zitadel_client_secret`, затем установить `AUTH_ZITADEL_ENABLED=true`
8. При необходимости включить browser return URL через `APPLICATION_PUBLIC_BASE_URL=http://api.localhost:8080` и `AUTH_BROWSER_DEFAULT_RETURN_URL=/auth/callback`
9. Если нужен platform endpoint `POST /v1/platform/users/{userId}/email/force-verify`, создать отдельный service account в ZITADEL, выдать ему admin-права и сохранить его PAT в `AUTH_ZITADEL_ADMIN_TOKEN` или `AUTH_ZITADEL_ADMIN_TOKEN_FILE`
10. Перезапустить `api`
11. После этого можно прогнать живой browser smoke: `make smoke-auth-zitadel-e2e`

Что важно в текущем auth-контуре:

- browser login через ZITADEL использует Authorization Code Flow + PKCE `S256`
- локальные `POST /v1/accounts` и `POST /v1/auth/login` остаются legacy fallback и управляются `AUTH_LOCAL_SIGNUP_ENABLED` и `AUTH_PASSWORD_LOGIN_ENABLED`
- liveness и readiness разведены: `GET /v1/health` и `GET /v1/ready`

## Contracts и snapshots

Для OpenAPI/router parity добавлена отдельная утилита:

```bash
go -C platform run ./cmd/contracts openapi-json
go -C platform run ./cmd/contracts openapi-yaml
go -C platform run ./cmd/contracts routes
go -C platform run ./cmd/contracts check-parity
```

Локальные удобные алиасы:

```bash
make check-contracts
make contracts-openapi-json
make contracts-openapi-yaml
make contracts-routes
make contracts-snapshot
```

Если текущая shell в WSL/OneDrive потеряла рабочий каталог и обычный `make` падает с `getcwd: No such file or directory`, можно использовать wrapper из корня репозитория:

```bash
bash /mnt/c/Users/nokclock/OneDrive/Документы/GitHub/collabsphere/makew migrations-build
```

Он запускает `make` через явный `-C <repo-root>` и не зависит от сломанного текущего `cwd`.

`make check-contracts` и `make contracts-*` используют встроенный минимальный dev-конфиг для `cmd/contracts`, поэтому не требуют заранее экспортировать `POSTGRES_*`, `AUTH_*` и другие runtime secrets. При необходимости app-level поля всё равно можно переопределить через обычные env overrides.

Это поведение теперь зафиксировано профильной валидацией:

- `api` валидирует полный runtime-контур, включая DB, JWT, browser auth, ZITADEL и включённые интеграции
- `worker` валидирует DB, JWT и только реально используемые worker-интеграции
- `migrate` и `seed` валидируют только DB-oriented subset и не требуют `AUTH_JWT_SECRET(_FILE)`
- `contracts` требует только app metadata и не зависит от DB/JWT runtime secrets

Команда `make contracts-snapshot` обновляет:

- `docs/openapi/openapi.json`
- `docs/openapi/openapi.yaml`
- `docs/openapi/routes.txt`

Если меняется behavior HTTP handlers, но snapshots не обновлены, это считается ошибкой контракта.

## Readiness и smoke

`GET /v1/health` — лёгкий liveness probe.

`GET /v1/ready` — readiness probe, который в текущем MVP проверяет доступность primary PostgreSQL connection.

Локальные smoke-скрипты не меняют окружение и рассчитаны на уже запущенный API:

```bash
./scripts/smoke-bootstrap.sh
./scripts/smoke-auth-legacy.sh
./scripts/smoke-auth-zitadel-e2e.sh
```

Или через `make`:

```bash
make smoke-bootstrap
make smoke-auth-legacy
make smoke-auth-zitadel-e2e
```

По умолчанию они бьют в `http://127.0.0.1:8080`. При необходимости можно переопределить `BASE_URL`.

`smoke-auth-zitadel-e2e` дополнительно требует:

- `Node/npm` только как test-only prerequisite, потому что browser automation идёт через Playwright CLI wrapper
- уже настроенный локальный ZITADEL OIDC application для backend callback
- рабочий `AUTH_ZITADEL_ADMIN_TOKEN` или `deploy/secrets/identity/zitadel_admin_token`

Сценарий использует seeded admin из `deploy/secrets/identity/zitadel_init_steps.yaml`, сам создаёт отдельного `unverified` пользователя через ZITADEL User API, подтверждает, что первый external login отклоняется, вызывает backend `force-verify`, а затем подтверждает успешный browser login и visible `accessToken` / `refreshToken` на `/auth/callback`. Артефакты браузера складываются в `output/playwright/`.

## Memberships и invitations

Организационный access-core в этом цикле усилен вокруг существующего tenant boundary `Organization`, без ввода отдельного `Workspace`.

Новые маршруты:

- `POST /v1/organizations/{organization_id}/invitations`
- `GET /v1/organizations/{organization_id}/invitations`
- `POST /v1/invitations/{token}/accept`

Что они делают:

- owners/admins могут выпустить time-limited invitation для email + target role
- invitations видны в organization-scoped list
- acceptance требует аутентифицированный account, и email приглашения должен совпадать с email текущего аккаунта
- mutating membership/invitation операции пишут `iam.organization_access_audit_events`

## CI

В репозитории добавлен baseline workflow:

- `.github/workflows/ci.yaml`

Он проверяет:

- core unit/compile tests
- bundled migrations parity
- OpenAPI/router parity
- contract snapshots
- bootstrap smoke
- legacy auth smoke
- DB-backed integration suites
- access token в backend - JWT, а refresh token - opaque session token
- refresh token rotation одноразовая: повторное использование уже ротированного refresh token отзывает всю refresh session

Для `deploy/secrets/identity/zitadel_master_key` используйте случайный ASCII-ключ ровно на 32 символа. В PowerShell его можно сгенерировать так:

```powershell
-join ((48..57 + 65..90 + 97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

`AUTH_ZITADEL_CLIENT_ID` и `deploy/secrets/identity/zitadel_client_secret` не генерируются заранее. Они появляются только после первого запуска ZITADEL, когда вы создаёте OIDC application в админ-панели и копируете оттуда выданные значения.

### ZITADEL admin token для force-verify

Чтобы `POST /v1/platform/users/{userId}/email/force-verify` реально работал, нужны две независимые авторизации:

1. server-side PAT, которым backend сам ходит в ZITADEL
2. backend access token уже авторизованного CollabSphere-пользователя с ролью `platform_admin`

Это control-plane endpoint. Он не предназначен для ситуации "у меня вообще нет рабочего admin-сеанса, но я хочу этим же запросом починить вход самому себе".

`AUTH_ZITADEL_ADMIN_TOKEN` не должен содержать `client_secret`, пароль пользователя или JSON-ответ OAuth. Здесь ожидается сырой Bearer token, и для текущего backend-кода правильный вариант - Personal Access Token сервисного аккаунта ZITADEL.

#### 1. Создать service account в ZITADEL

1. Открыть консоль ZITADEL: [http://auth.localhost:8090/ui/console](http://auth.localhost:8090/ui/console)
2. Перейти в `Users -> Service Accounts`
3. Создать service account
   Рекомендуемые поля:
   `User Name`: `collabsphere-platform-service`
   `Name`: `CollabSphere Platform Service`
   `Access Token Type`: `Bearer`
4. Открыть созданный service account и создать для него Personal Access Token
5. Сразу сохранить показанное значение: ZITADEL показывает PAT только один раз
6. Выдать service account административную роль
   Для локального org-scoped dev обычно достаточно назначить его администратором организации `CollabSphere` с ролью `Org Owner`
   Если backend потом отвечает `PLATFORM_ZITADEL_UNAVAILABLE` или ZITADEL возвращает `401/403`, повысить права до instance-level `IAM_OWNER`

#### 2. Подключить PAT к backend

Сохранить токен в `deploy/env/.env.zitadel.dev`:

```env
AUTH_ZITADEL_ADMIN_TOKEN=<PASTE_PAT_HERE>
```

Или сохранить токен в файл и указать file-based режим:

```env
AUTH_ZITADEL_ADMIN_TOKEN_FILE=/run/secrets/zitadel_admin_token
AUTH_ZITADEL_ADMIN_TOKEN_SOURCE_FILE=secrets/identity/zitadel_admin_token
```

По умолчанию file-based режим выключен: `AUTH_ZITADEL_ADMIN_TOKEN_FILE` пустой, а Compose подставляет безопасный placeholder secret, чтобы `api` не падал на старте. Чтобы включить этот режим, создайте `deploy/secrets/identity/zitadel_admin_token`, запишите туда raw PAT и установите значения выше. Docker Compose прочитает этот host-файл и смонтирует его в контейнер `api` как `/run/secrets/zitadel_admin_token`. Содержимое файла должно быть только токеном, без JSON-обёртки.

После изменения PAT нужно пересоздать `api`:

```bash
docker compose \
  --env-file deploy/env/.env.dev \
  --env-file deploy/env/.env.postgres.dev \
  --env-file deploy/env/.env.storage.dev \
  --env-file deploy/env/.env.redis.dev \
  --env-file deploy/env/.env.zitadel.dev \
  -f deploy/compose/core.yaml \
  -f deploy/compose/storage.yaml \
  -f deploy/compose/auth.yaml \
  --profile local up -d --build --force-recreate api
```

#### 3. Подготовить вызывающий backend access token

Нужен отдельный токен CollabSphere-пользователя, который уже может входить в backend и имеет effective role `platform_admin`.

Проверка:

1. Получить обычный backend access token любым рабочим способом
   `POST /v1/auth/login`
   или browser login через ZITADEL + `POST /v1/auth/exchange`
1. Проверить, какой локальный account используется этим токеном:

```bash
curl -H "Authorization: Bearer <backend_access_token>" \
  http://api.localhost:8080/v1/auth/me
```

1. Проверить control-plane access:

```bash
curl -H "Authorization: Bearer <backend_access_token>" \
  http://api.localhost:8080/v1/platform/access/me
```

Ожидается, что в ответе `effectiveRoles` содержит `platform_admin`.

Если здесь `403 Platform access denied`, а это локальный dev, самый прямой способ разблокироваться:

1. Взять локальный `id` из ответа `GET /v1/auth/me`
2. Добавить этот UUID в `AUTH_PLATFORM_BOOTSTRAP_ACCOUNT_IDS` в `deploy/env/.env.dev`
3. Пересоздать `api`
4. Повторить `GET /v1/platform/access/me`

#### 4. Проверить PAT напрямую против ZITADEL

Перед вызовом backend route полезно убедиться, что server-side PAT действительно может читать целевого пользователя:

```bash
curl -H "Authorization: Bearer $(tr -d '\r\n' < deploy/secrets/identity/zitadel_admin_token)" \
  http://auth.localhost:8090/v2/users/<zitadel_user_id>
```

Если эта проверка даёт auth/permission error, сама ручка `force-verify` тоже не сработает.

#### 5. Выполнить force-verify через backend route

Где взять `userId`:

- это ZITADEL user id, а не локальный `account_id`
- его можно взять в ZITADEL console или через `GET /v2/users/{id}`

Вызов:

```http
POST /v1/platform/users/{userId}/email/force-verify
Authorization: Bearer <backend access token>
```

Что важно:

- backend-to-ZITADEL credential и пользовательский токен backend - разные вещи
- `AUTH_ZITADEL_ADMIN_TOKEN` использует только backend
- без этого токена route не должен выполнять verification и будет возвращать `PLATFORM_ZITADEL_ADMIN_DISABLED`
- целевого пользователя верифицирует уже backend через свой service-account PAT, а не ваш пользовательский bearer token

На текущем backend flow делает так:

1. читает пользователя из ZITADEL
2. пытается переиспользовать существующий email verification code через `email/resend`
3. если существующего кода нет, запрашивает новый через `email/send`
4. сразу подтверждает email через `email/verify`

#### 6. Закрыть happy-path живым browser smoke

Когда backend OIDC app и admin PAT уже настроены, этот сценарий можно проверить end-to-end одной командой:

```bash
make smoke-auth-zitadel-e2e
```

Что делает smoke:

1. логинится в browser flow как seeded admin `admin@collabsphere.ru`
2. забирает backend `accessToken` прямо с dev callback page `/auth/callback`
3. подтверждает `platform_admin` через `GET /v1/platform/access/me`
4. создаёт нового `unverified` ZITADEL user через User API
5. доказывает, что первый external login этого user отклоняется с `Verified email is required for first external login`
6. вызывает `POST /v1/platform/users/{userId}/email/force-verify`
7. повторяет browser login тем же user и подтверждает, что callback page показывает локальные `accessToken` и `refreshToken`

Если в локальной системе нет `npx`, smoke завершится с явной подсказкой по установке Node/npm и Playwright CLI.

Если вы всё ещё видите ответ вида `Code is empty (EMAIL-...)`, это сильное свидетельство, что запущен старый `api` image без этого fallback. В таком случае нужно именно пересобрать и пересоздать `api`, а не только перезапустить контейнер.

Подробная пошаговая инструкция вынесена в [`docs/content/authentication/zitadel.md`](docs/content/authentication/zitadel.md).

По умолчанию используется hostname `auth.localhost`.
Это сделано намеренно: браузер на хосте должен открывать `http://auth.localhost:8090` и `http://auth.localhost:3000`, а контейнер `api` для server-side OIDC discovery и token exchange тоже должен использовать `http://auth.localhost:8090` через `host-gateway`. Использование `auth.localhost:8080` внутри `api` в текущей compose-схеме уводит запросы обратно в опубликованный порт самого `api`, а использование `http://zitadel:8080` ломает instance resolution в ZITADEL, потому что инстанс зарегистрирован на `auth.localhost`, а не на `zitadel`.
`deploy/secrets/identity/zitadel_init_steps.yaml` применяется только на первом bootstrap инстанса. Если ZITADEL уже инициализирован, изменение этого файла само по себе не обновит существующего администратора.

### Остановка окружения

```bash
docker compose \
  -f deploy/compose/core.yaml \
  -f deploy/compose/storage.yaml \
  -f deploy/compose/auth.yaml \
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
- `deploy/secrets/jwt/jwt_secret_key`
- `deploy/secrets/identity/zitadel_client_secret`
- `deploy/secrets/identity/zitadel_master_key`
- `deploy/secrets/identity/zitadel_runtime_secrets.yaml`
- `deploy/secrets/identity/zitadel_init_steps.yaml`

Для локального ZITADEL используется схема с файловыми секретами:

- `zitadel_master_key` монтируется в контейнер и передаётся через `--masterkeyFile`
- `zitadel_runtime_secrets.yaml` подключается как приватный `--config` с паролями PostgreSQL
- `zitadel_init_steps.yaml` подключается как приватный `--steps` с bootstrap-настройками первого администратора и login client
- `zitadel_client_secret` монтируется в `api`, а Go runtime читает его через `AUTH_ZITADEL_CLIENT_SECRET_FILE`

Даже для локального профиля часть секретов объявлена на уровне сервиса, поэтому отсутствие этих файлов ломает запуск Compose.

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

`make migrate-up` и `make migrate-down` используют `deploy/compose/jobs.yaml` и запускают контейнер с `platform/cmd/migrate`.

Перед запуском `make migrate-*` и `make seed-*` теперь пересобирают локальный образ `colabsphere-api:${IMAGE_TAG}`, чтобы мигратор и сидер всегда поднимались на актуальном коде, а не на старом ранее собранном образе.

После профильного refactor'а `platform/cmd/migrate` и `platform/cmd/seed` больше не зависят от `AUTH_JWT_SECRET(_FILE)` или browser/OIDC secrets. Для них по-прежнему обязателен только валидный `POSTGRES_*` набор.

Если вы меняете SQL в `migrations-src/`, сначала пересоберите bundle через `make migrations-build`, а уже потом прогоняйте миграции. `make check-migrations` теперь сравнивает bundle до и после rebuild через обычный `diff`, так что для него не нужен `git`.

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

Точечный прогон auth и OIDC тестов:

```bash
go -C platform test ./internal/auth/... ./internal/runtime/infrastructure/security/oidc
```

Integration-тесты auth legacy login / refresh / logout / me:

```bash
go -C platform test -tags=integration ./internal/auth/delivery/http -run 'TestLegacyPasswordLoginIntegration|TestAuthMeIntegration'
```

Integration-тесты создания аккаунта:

```bash
go -C platform test -tags=integration ./internal/accounts/delivery/http -run 'TestCreateAccountIntegration'
```

Integration-тесты организаций:

```bash
go -C platform test -tags=integration ./internal/organizations/delivery/http -run 'Test.*Organization.*Integration'
```

Integration-тесты platform organization review:

```bash
go -C platform test -tags=integration ./internal/platformops/delivery/http -run 'Test.*Review.*Integration'
```

Integration-тесты memberships:

```bash
go -C platform test -tags=integration ./internal/memberships/delivery/http -run 'TestMembershipsIntegration'
```

Через `make` этот набор можно запускать так:

```bash
make test-platform-reviews-integration
```

Для integration-прогона нужен `COLLABSPHERE_TEST_POSTGRES_DSN`. Если переменная не задана, тесты этих наборов будут корректно `SKIP`.

Если хотите гонять эти наборы через отдельный test-only PostgreSQL, последовательность такая:

```bash
docker compose --env-file deploy/env/.env.postgres.test -f deploy/compose/test.yaml --profile test up -d postgres-test
export COLLABSPHERE_TEST_POSTGRES_DSN="host=127.0.0.1 port=5434 user=postgres password=$(cat deploy/secrets/postgres/test/db_password) dbname=postgres sslmode=disable"
```

Здесь DSN должен смотреть именно на базу `postgres`, а не на `collabsphere_test`, потому что integration-тесты создают и удаляют временные базы через `CREATE DATABASE`.

Это же полезно запускать перед сборкой Docker-образа. В `platform/Dockerfile` дополнительно выполняются `go vet ./...` и `go test -v ./...` на build stage.

## API и документация

### Базовый префикс

Все зарегистрированные маршруты API живут под префиксом `/v1`.

### Redirects из корня

Корневой router дополнительно пробрасывает удобные entrypoints:

- `/openapi.yaml` -> `/v1/openapi.yaml`
- `/health` -> `/v1/health`

### Основные группы маршрутов

Системные:

- `GET /v1/health`

Accounts:

- `POST /v1/accounts`
- `GET /v1/accounts/{id}`
- `GET /v1/accounts/by-email`

Organizations:

- `POST /v1/organizations`
- `GET /v1/organizations/{id}`

Memberships:

- `POST /v1/organizations/{organization_id}/members`
- `GET /v1/organizations/{organization_id}/members`

Auth:

- `POST /v1/auth/login`
- `GET /v1/auth/zitadel/login`
- `GET /v1/auth/zitadel/signup`
- `GET /v1/auth/zitadel/callback`
- `POST /v1/auth/exchange`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `GET /v1/auth/me`

Auth-маршруты регистрируются в bootstrap и используют реальный JWT token manager. Legacy password login и local signup остаются флагируемыми fallback-сценариями, а browser login через ZITADEL зависит от корректной внешней конфигурации OIDC и секретов.

`web/` не является вторым backend и не дублирует auth/business logic. Он работает поверх тех же `/v1/auth/*`, `/v1/organizations/*` и связанных runtime-маршрутов.

## Known issues

- Старый корневой README был неполным и частично расходился с текущим кодом и Compose-конфигурацией. Этот документ заменяет его целиком.
- `Makefile` ориентирован на Linux / WSL и использует `bash`. В PowerShell и CMD лучше вызывать `docker compose` напрямую.
- В `deploy/.env` есть расхождения с текущим runtime-кодом, включая строку `AUTH_JWT_SECRET_FILE==...`. Кроме того, часть переменных из файла сейчас не читается приложением вообще.
- Compose-конфигурация и env уже содержат задел под будущие подсистемы и внешние интеграции. Не вся эта конфигурация соответствует текущему фактическому поведению API.
- Для ZITADEL browser-flow добавлен отдельный live smoke `scripts/smoke-auth-zitadel-e2e.sh` и отдельный CI job `zitadel-e2e`, но этот gate сейчас честно зависит от уже provisioned local OIDC app и admin PAT. На bare checkout без live ZITADEL secret files job пропускается, а не делает вид, что может доказать live E2E.
- Локальный observability-стек `Grafana + Loki + Alloy` добавлен отдельно от основного compose-стека и требует рабочего локального Docker runtime; container-level валидация `Alloy` может упираться в локальные Docker credential/pull проблемы.
- `web/` добавлен как рабочий frontend shell, но в этой среде я не смог прогнать `npm install`, `npm run lint`, `npm run typecheck` и `npm run build`, потому что локально отсутствуют `node`, `npm` и `npx`. Поэтому frontend skeleton проверен статически, но не реальным Next build.
- В текущем `web/` токены временно хранятся в `localStorage`. Это осознанный MVP-компромисс, а не финальная security-модель.

## Полезные ссылки

- ADR по архитектурным границам: [`docs/architecture/adr-foundation-infrastructure-boundaries.md`](docs/architecture/adr-foundation-infrastructure-boundaries.md)
- Compose-файлы: `deploy/compose/core.yaml`, `deploy/compose/jobs.yaml`
- Исходники API: `platform/cmd/api`, `platform/internal/`
