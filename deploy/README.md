# Deploy Layout

`deploy/` intentionally keeps compose entrypoints at the top level, while stack-specific assets live in subdirectories.

## Structure

- `deploy/docker-compose.*.yaml`: compose entrypoints for app, db, migrations, storage, identity, web and observability
- `deploy/docker-compose.web.yaml`: отдельный compose-layer для `web/` frontend поверх основного platform stack
- `deploy/env/`: local env files such as `.env.dev`, `.env.postgres.dev`, `.env.postgres.test`, `.env.example`
- `deploy/env/.env.postgres.dev`: PostgreSQL overlay for the main platform stack
- `deploy/env/.env.postgres.test`: PostgreSQL overlay for integration/test database
- `deploy/env/.env.storage.dev`: storage/S3 overlay
- `deploy/env/.env.redis.dev`: realtime/redis overlay
- `deploy/env/.env.zitadel.dev`: ZITADEL/OIDC overlay
- `deploy/env/.env.web.dev`: отдельный frontend overlay для `WEB_*`
- `deploy/observability/`: Grafana, Loki, Alloy and Prometheus configs
- `deploy/scripts/`: helper scripts used by compose stacks
- `deploy/secrets/`: local file-based secrets for dev/test stacks
- `deploy/fixtures/`: disabled placeholders and bootstrap fixtures

## Notes

- Main local stack commands should now use `deploy/env/.env.dev`
- Main local stack commands should also include `deploy/env/.env.postgres.dev`
- Storage-aware compose commands should also include `deploy/env/.env.storage.dev`
- Redis-aware compose commands should also include `deploy/env/.env.redis.dev`
- ZITADEL-aware compose commands should also include `deploy/env/.env.zitadel.dev`
- Frontend container stack composes on top of the main local stack via `deploy/docker-compose.web.yaml`
- Frontend compose commands should combine base + postgres + storage + redis + zitadel + web overlays
- Test postgres stack should use `deploy/env/.env.postgres.test`
- Observability configs now live under `deploy/observability/*`
