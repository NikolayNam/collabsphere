# CollabSphere Web

`web/` - это frontend shell на `Next.js`, работающий поверх текущего Go backend из [`platform/`](../platform/).

Он не является вторым backend и не дублирует доменную логику. Source of truth для auth, organizations, memberships и control-plane остаётся в `/v1` API.

## Что внутри

- `Next.js` с `App Router`
- `TypeScript`
- `ESLint`
- frontend auth helpers поверх существующих backend routes
- Next rewrites на `/api/backend/*`, чтобы не тащить CORS-переделку в первый же frontend-коммит

## Текущие страницы

- `/` - overview и app shell
- `/login` - ZITADEL browser login и legacy fallback login
- `/auth/callback` - `ticket -> /v1/auth/exchange`
- `/me` - `GET /v1/auth/me`, refresh и logout
- `/organizations` - create organization, profile editor, logo upload и resolve-by-host
- `/chat` - groups -> channels -> messages поверх существующих collab endpoints

## Требования

- `Node.js >= 20.9`
- `npm`
- уже запущенный backend API

Рекомендуемые локальные адреса:

- frontend: `http://collabsphere.localhost:3002`
- backend: `http://api.localhost:8080`

## Быстрый старт

Из корня репозитория:

```bash
cd web
cp .env.local.example .env.local
npm install
npm run dev
```

После этого frontend будет доступен на:

- [http://collabsphere.localhost:3002](http://collabsphere.localhost:3002)

## Docker Compose запуск

Если хотите поднять frontend контейнером вместе с backend-стеком:

```bash
make up-dev-web
```

Этот путь использует:

- `deploy/env/.env.dev` как базовый env для platform stack
- `deploy/env/.env.postgres.dev` как PostgreSQL overlay
- `deploy/env/.env.storage.dev` для storage/S3
- `deploy/env/.env.redis.dev` для realtime/redis
- `deploy/env/.env.zitadel.dev` для ZITADEL/OIDC
- `deploy/env/.env.web.dev` как overlay для `WEB_*`

Или напрямую:

```bash
docker compose \
  --env-file deploy/env/.env.dev \
  --env-file deploy/env/.env.postgres.dev \
  --env-file deploy/env/.env.storage.dev \
  --env-file deploy/env/.env.redis.dev \
  --env-file deploy/env/.env.zitadel.dev \
  --env-file deploy/env/.env.web.dev \
  -f deploy/compose/core.yaml \
  -f deploy/compose/storage.yaml \
  -f deploy/compose/auth.yaml \
  --profile local --profile web up -d --build --force-recreate
```

По умолчанию containerized frontend слушает `http://localhost:3002`, а для browser auth flow использует внешний origin `http://collabsphere.localhost:3002`. Если у вас `collabsphere.localhost` не резолвится в loopback, можно временно переопределить `WEB_NEXT_PUBLIC_APP_BASE_URL=http://localhost:3002` и синхронно обновить `WEB_AUTH_BROWSER_REDIRECT_ORIGINS`.

Важно: `deploy/compose/core.yaml` содержит web-профиль и дополняет `api`-конфиг значением `AUTH_BROWSER_REDIRECT_ORIGINS`, чтобы combined stack принимал frontend callback origin без ручного редактирования backend env.

## Переменные окружения

См. [`.env.local.example`](./.env.local.example):

```env
NEXT_PUBLIC_API_BASE_URL=http://api.localhost:8080
NEXT_PUBLIC_APP_BASE_URL=http://collabsphere.localhost:3002
NEXT_INTERNAL_API_BASE_URL=http://api.localhost:8080
```

Что они значат:

- `NEXT_PUBLIC_API_BASE_URL` — реальный backend origin
- `NEXT_PUBLIC_APP_BASE_URL` — внешний frontend origin, который backend использует как `return_to` для browser auth flow
- `NEXT_INTERNAL_API_BASE_URL` — backend origin для Next rewrites; при локальном `npm run dev` обычно совпадает с `NEXT_PUBLIC_API_BASE_URL`, а в Docker Compose это `http://api:8080`

## Как работает auth

Frontend использует уже существующий backend flow:

1. `/login` отправляет браузер на:
   - `GET /v1/auth/zitadel/login`
   - или `GET /v1/auth/zitadel/signup`
2. backend переводит пользователя в self-hosted login UI на `auth.localhost:3000`
3. после OIDC callback backend редиректит пользователя на frontend `auth/callback` с `?ticket=...`
4. frontend вызывает `POST /v1/auth/exchange`
5. backend возвращает локальные `accessToken` / `refreshToken`
6. frontend использует их для `GET /v1/auth/me`, `POST /v1/auth/refresh`, `POST /v1/auth/logout`

## Почему здесь есть rewrite proxy

Сейчас frontend ходит в backend через:

- `/api/backend/*`

Это проксируется на реальный backend в [`next.config.ts`](./next.config.ts).

Причина простая:

- в текущем Go runtime нет отдельного выделенного CORS-контура
- так проще получить рабочий frontend без немедленного изменения backend transport-layer

Важно:

- browser navigation на `GET /v1/auth/zitadel/login` идёт напрямую в backend
- через rewrite идут только frontend XHR/fetch-запросы

## Полезные команды

```bash
npm run dev
npm run lint
npm run typecheck
npm run build
npm run start
```

## Текущие ограничения

- токены пока хранятся в `localStorage`
- это MVP-компромисс, а не финальная security-модель
- production-ready вариант лучше делать через `httpOnly` cookies / BFF слой
- chat пока обновляется через polling, а не через WebSocket realtime
- frontend пока не покрыт отдельным e2e/lint/typecheck прогоном в этой среде, потому что локально у меня нет `node`/`npm`
- containerized `web` build я тоже не прогонял end-to-end через реальный `docker build`; нижеописанный compose-контур проверяется конфигурационно

## Следующие шаги

- перевести токены из `localStorage` в `httpOnly` cookies
- решить постоянную стратегию `proxy vs CORS`
- добавить tenant screens и control-plane UI
- подключить frontend build/lint/typecheck в CI
