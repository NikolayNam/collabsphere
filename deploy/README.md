# Deploy Layout

`deploy/` intentionally keeps compose entrypoints at the top level, while stack-specific assets live in subdirectories.

## Structure

- `deploy/docker-compose.*.yaml`: compose entrypoints for app, db, migrations, storage, identity and observability
- `deploy/env/`: local env files such as `.env.dev`, `.env.test`, `.env.example`
- `deploy/observability/`: Grafana, Loki, Alloy and Prometheus configs
- `deploy/scripts/`: helper scripts used by compose stacks
- `deploy/secrets/`: local file-based secrets for dev/test stacks
- `deploy/fixtures/`: disabled placeholders and bootstrap fixtures

## Notes

- Main local stack commands should now use `deploy/env/.env.dev`
- Test postgres stack should use `deploy/env/.env.test`
- Observability configs now live under `deploy/observability/*`
