# Deploy Layout

`deploy/` keeps compose entrypoints grouped under `deploy/compose/`, while stack-specific assets live in subdirectories.

## Structure

- `deploy/compose/core.yaml`: core stack (api, postgres, optional web/web-login profile)
- `deploy/compose/auth.yaml`: identity stack (ZITADEL)
- `deploy/compose/storage.yaml`: storage stack (SeaweedFS)
- `deploy/compose/jobs.yaml`: one-shot jobs (migrate/seed)
- `deploy/compose/observability.yaml`: observability stack (Prometheus/Loki/Alloy/Grafana)
- `deploy/compose/test.yaml`: isolated test postgres stack
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
- Frontend container stack uses the `web` profile from `deploy/compose/core.yaml`
- Frontend compose commands should combine base + postgres + storage + redis + zitadel + web overlays
- Test postgres stack should use `deploy/env/.env.postgres.test`
- Observability configs now live under `deploy/observability/*`
