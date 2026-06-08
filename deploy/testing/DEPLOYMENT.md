# new-api 测试环境远程中间件部署文档

本文件用于 `new-api` 本地测试部署：应用运行在本机，连接远程测试环境 PostgreSQL 和 Redis。

## 部署信息

- 项目：`new-api`
- 运行方式：本地运行项目或本地容器运行项目
- PostgreSQL：`139.155.130.17:5432`
- Redis：`139.155.130.17:6379`
- PostgreSQL 测试库：`newapi`
- Redis 隔离库：DB `4`
- 真实中间件凭据来源：`/Users/mac/work/中间件/credentials.yaml`

## 测试环境约束

- 不得提交真实数据库密码、Redis 密码、连接串、token、`SESSION_SECRET` 或 `CRYPTO_SECRET`。
- 不得修改 `/Users/mac/work/中间件/credentials.yaml`。
- 不得在本地测试部署中启动 PostgreSQL、MySQL 或 Redis 容器。
- 不得把 `SQL_DSN` 指向 `manbotv`、`sub2api_test` 或其它已有业务数据库。
- 不得删除远程数据库，不得清空表，不得执行 `DROP DATABASE`、`DROP SCHEMA` 或全库 `TRUNCATE`。
- 不得执行 Redis `FLUSHDB`、`FLUSHALL`，不得清理、修改或写入 Redis DB `0/1/2/3` 或其它 DB。
- `REDIS_CONN_STRING` 必须以 `/4` 结尾，确保测试数据写入 Redis DB `4`。

## 远程中间件现状

已做非破坏性连通性检查：

```bash
nc -vz 139.155.130.17 5432
nc -vz 139.155.130.17 6379
```

当前远程 PostgreSQL 非模板数据库包括：

```text
manbotv
postgres
sub2api_test
```

当前检查结论：

- PostgreSQL 可连通。
- Redis 可连通。
- `newapi` 数据库当前不存在，因此尚未完成本项目测试数据库隔离。
- `manbotv` 已有表数据，不允许作为本项目测试数据库使用。

## PostgreSQL 初始化

首次部署前必须创建独立数据库 `newapi`。如果 `newapi` 已存在，不得删除、不得重建、不得清空数据，只允许应用启动后的 GORM migration 修改表结构。

使用 `credentials.yaml` 中的 PostgreSQL admin 凭据连接远程 PostgreSQL 后执行：

```sql
SELECT 'CREATE DATABASE newapi'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'newapi')\gexec
```

验证当前数据库：

```sql
SELECT current_database();
```

预期连接到 `newapi` 后输出：

```text
newapi
```

也可以检查数据库是否存在：

```sql
SELECT datname FROM pg_database WHERE datname = 'newapi';
```

## 本地环境变量

真实密码只允许从 `/Users/mac/work/中间件/credentials.yaml` 手工读取，并写入本机未提交的 `.env`。不要把真实值写入仓库文档或示例文件。

本地 `.env` 应使用以下形态：

```env
SQL_DSN=postgresql://admin:<password>@139.155.130.17:5432/newapi?sslmode=disable
REDIS_CONN_STRING=redis://admin:<password>@139.155.130.17:6379/4
CRYPTO_SECRET=<local-test-fixed-secret>
SESSION_SECRET=<local-test-session-secret>
TZ=Asia/Shanghai
ERROR_LOG_ENABLED=true
BATCH_UPDATE_ENABLED=true
NODE_NAME=new-api-test-local
```

启动前必须确认：

```bash
printf '%s\n' "$SQL_DSN" | grep '139.155.130.17:5432/newapi'
printf '%s\n' "$REDIS_CONN_STRING" | grep '/4$'
```

## 启动与验证

启动前检查远程端口：

```bash
nc -vz 139.155.130.17 5432
nc -vz 139.155.130.17 6379
```

验证 Redis DB `4` 可用：

```bash
redis-cli -h 139.155.130.17 -p 6379 -a '<password>' -n 4 ping
```

预期输出：

```text
PONG
```

启动 `new-api` 后检查：

```bash
curl -fsS http://127.0.0.1:3000/api/status
```

日志验收标准：

- 必须显示使用 PostgreSQL。
- 不得出现 SQLite fallback。
- 必须显示 Redis enabled。

## 验收标准

- PostgreSQL 中存在独立数据库 `newapi`。
- 本地 `.env` 的 `SQL_DSN` 指向 `139.155.130.17:5432/newapi`。
- 本地 `.env` 的 `REDIS_CONN_STRING` 以 `/4` 结尾。
- 应用启动后使用 PostgreSQL，不回退 SQLite。
- 应用测试数据只写入 PostgreSQL `newapi` 和 Redis DB `4`。
