# AGENTS.md - Production Deployment Rules

This directory contains production deployment configuration for `new-api`.

## Production Safety

- Treat every file in this directory as production-impacting.
- Do not add PostgreSQL, MySQL, or Redis services to `docker-compose.yml`.
- Production images must be pushed to both Docker Hub `asd494235908/new-api` and Aliyun `registry.cn-chengdu.aliyuncs.com/slef/new-api`.
- Both image registries must use the same immutable tag, for example `prod-YYYYMMDD-<git-short-sha>-<timestamp>-amd64`; do not use `latest` or `local` in production.
- Before deployment, confirm both Docker Hub and Aliyun image manifests exist and include the `linux/amd64` platform.
- Production deployment targets are server A `82.156.54.174` and server B `43.160.242.244`.
- Deploy `new-api` to both server A and server B unless the operator explicitly requests a single-server deployment.
- Do not modify `/opt/services/sub2api/`.
- On both production servers, only `new-api` may be recreated or restarted during deployment.
- Do not stop, restart, or recreate `sub2api`, `sub2api-postgres`, `sub2api-redis`, or any non-`new-api` container.
- Do not change existing `sub2api` containers, env files, data volumes, or Docker networks unless explicitly instructed by the operator.
- On `43.160.242.244`, connect to existing PostgreSQL and Redis through `host.docker.internal` mapped to Docker `host-gateway`; do not assume Docker service DNS names resolve.

## PostgreSQL Rules

- Server A and server B `new-api` must both connect to server A's isolated `new_api` database.
- Server B must not use its local `new_api` database as the formal production database.
- Never point `SQL_DSN` at the `sub2api` database.
- Create a dedicated `new_api` PostgreSQL user for `new-api`.
- Grant the `new_api` user access only to the `new_api` database.
- Do not run `new-api` migrations against the `sub2api` database.
- Keep conservative connection pool limits to avoid starving `sub2api`.

## Redis Rules

- `new-api` must use the existing `sub2api-redis` service with Redis DB `1`.
- Never use Redis DB `0`; it is reserved for `sub2api`.
- Never run `FLUSHDB`, `FLUSHALL`, or any bulk key deletion against `sub2api-redis`.
- Do not change the `sub2api-redis` password or service configuration.

## Secrets

- Never commit real database passwords, Redis passwords, SSH passwords, API keys, or `SESSION_SECRET`.
- Store real production values only in `/opt/services/new-api/.env` on the server.
- Keep `.env.example` as placeholders only.

## Required Checks

Before deployment:

- Confirm `sub2api`, `sub2api-postgres`, and `sub2api-redis` are running.
- Confirm Docker network `sub2api-migrated` exists.
- Confirm port `3000` is not already listening.
- Confirm `SQL_DSN` points to `new_api`, not `sub2api`.
- Confirm `REDIS_CONN_STRING` ends with `/1`.
- Confirm `MODEL_SQUARE_ENVIRONMENT` is explicitly set to `overseas` for overseas production or `domestic` for domestic production.

After deployment:

- Check `http://127.0.0.1:3000/api/status`.
- Check `http://43.160.242.244:3000/api/status`.
- Confirm `sub2api` remains healthy.
- Confirm `new-api` logs show PostgreSQL usage and do not show SQLite fallback.
