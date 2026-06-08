# new-api 生产部署文档

本目录保存 `new-api` 在服务器 `43.160.242.244` 上的生产部署配置。

## 部署信息

- 服务器：`43.160.242.244`
- SSH 用户：`ubuntu`
- 服务器部署目录：`/opt/services/new-api`
- 公网访问地址：`http://43.160.242.244:3000`
- Compose 文件：`/opt/services/new-api/docker-compose.yml`
- 环境变量文件：`/opt/services/new-api/.env`
- 现有 Docker 网络：`sub2api-migrated`
- 运行时数据库/Redis 访问方式：通过 Docker `host-gateway` 映射的 `host.docker.internal`

## 隔离规则

- 本部署不得启动新的 PostgreSQL、MySQL 或 Redis 容器。
- 复用现有 `sub2api-postgres` 服务，但必须创建并使用独立的 `new_api` 数据库。
- 创建独立的 PostgreSQL 用户 `new_api`，并且只授权访问 `new_api` 数据库。
- `new-api` 禁止连接现有 `sub2api` 数据库。
- 复用现有 `sub2api-redis` 服务，但必须使用 Redis DB `1`。
- Redis DB `0` 保留给 `sub2api` 使用，`new-api` 禁止使用 DB `0`。
- 禁止修改 `/opt/services/sub2api/`、`sub2api` 的 compose 文件、env 文件、容器和数据卷。
- 部署过程中禁止停止或重启 `sub2api`、`sub2api-postgres`、`sub2api-redis`。

## 部署前检查

部署前在服务器执行：

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
docker network inspect sub2api-migrated >/dev/null
docker exec sub2api-postgres pg_isready -U sub2api -d sub2api
SUB2API_REDIS_PASSWORD="$(docker exec sub2api sh -lc 'printf "%s" "$REDIS_PASSWORD"')"
docker exec sub2api-redis redis-cli -a "$SUB2API_REDIS_PASSWORD" -n 0 ping
docker exec sub2api-redis redis-cli -a "$SUB2API_REDIS_PASSWORD" -n 1 ping
ss -ltn | awk 'NR==1 || /:3000 /'
```

启动 `new-api` 前，`3000` 端口检查应没有监听进程。

在 `43.160.242.244` 上，`sub2api-postgres` 和 `sub2api-redis` 是通过 host 网络访问的，不能从 `new-api` 容器内依赖 Docker 服务名解析。生产 compose 因此将 `host.docker.internal` 映射到 Docker `host-gateway`。服务器 `/opt/services/new-api/.env` 应使用：

```env
SQL_DSN=postgresql://new_api:<password>@host.docker.internal:5432/new_api?sslmode=disable
REDIS_CONN_STRING=redis://:<password>@host.docker.internal:6379/1
```

## PostgreSQL 初始化

在现有 `sub2api-postgres` 服务内创建隔离数据库和用户。密码应随机生成，并且只保存到服务器 `/opt/services/new-api/.env`。

```bash
NEW_API_DB_PASSWORD='<generated-password>'

docker exec -i sub2api-postgres psql -U sub2api -d postgres <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'new_api') THEN
    CREATE ROLE new_api LOGIN PASSWORD '${NEW_API_DB_PASSWORD}';
  ELSE
    ALTER ROLE new_api WITH PASSWORD '${NEW_API_DB_PASSWORD}';
  END IF;
END
\$\$;
SQL

docker exec -i sub2api-postgres psql -U sub2api -d postgres <<SQL
SELECT 'CREATE DATABASE new_api OWNER new_api'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'new_api')\gexec
SQL

docker exec -i sub2api-postgres psql -U sub2api -d new_api <<SQL
REVOKE ALL ON DATABASE new_api FROM PUBLIC;
GRANT CONNECT ON DATABASE new_api TO new_api;
GRANT USAGE, CREATE ON SCHEMA public TO new_api;
ALTER SCHEMA public OWNER TO new_api;
SQL
```

验证新账号：

```bash
docker exec -e PGPASSWORD="$NEW_API_DB_PASSWORD" sub2api-postgres \
  psql -h 127.0.0.1 -U new_api -d new_api -tAc 'select current_database(), current_user'
```

预期输出包含 `new_api|new_api`。

## 构建与启动

本次生产部署不在服务器保留源码。推荐流程是：本地构建前端资源并交叉编译 `linux/amd64` 二进制，然后只上传运行包和部署配置到服务器，最后在服务器上构建运行镜像。

服务器目录应只保留运行包、compose、`.env`、文档、`data/` 和 `logs/`，不要保留 `/opt/services/new-api/src` 源码目录。

```bash
cd /opt/services/new-api
mkdir -p data logs
cp .env.example .env  # 仅首次部署时执行；已有生产 .env 时不要覆盖
```

在服务器编辑 `.env` 并替换所有占位符。真实密钥只允许保存在服务器 `.env` 中：

```bash
nano .env
```

本次已部署生产镜像 tag：

```env
NEW_API_IMAGE=new-api:prod-20260605-b0ac0429-amd64
```

生产镜像 tag 必须使用不可变命名，推荐格式：

```text
new-api:prod-YYYYMMDD-<git-short-sha>-amd64
```

不要在生产 `.env` 中使用 `latest`、`local` 或其他会随时间变化的 tag。

启动服务：

```bash
docker compose up -d
docker compose ps
docker compose logs -f new-api
```

## 健康检查

部署后执行：

```bash
curl -fsS http://127.0.0.1:3000/api/status
curl -fsS http://43.160.242.244:3000/api/status
docker ps --filter name=sub2api --format 'table {{.Names}}\t{{.Status}}'
docker compose logs new-api | grep -E 'PostgreSQL|using PostgreSQL|SQL_DSN|SQLite' || true
```

`new-api` 日志必须显示使用 PostgreSQL，不能出现 SQLite fallback。

## 日志

```bash
cd /opt/services/new-api
docker compose logs -f new-api
tail -f logs/*.log
```

## 更新

```bash
cd /opt/services/new-api
docker compose pull new-api || true
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

如果使用本地构建镜像，先在 `.env` 中更新 `NEW_API_IMAGE` 为新的镜像 tag，再执行 `docker compose up -d`。

## 回滚

将 `/opt/services/new-api/.env` 中的 `NEW_API_IMAGE` 改回上一个可用镜像 tag，然后只重启 `new-api`：

```bash
cd /opt/services/new-api
docker compose up -d --no-deps new-api
docker compose logs -f new-api
```

回滚过程中不要重启 `sub2api`、`sub2api-postgres` 或 `sub2api-redis`。

## 2026-06-05 部署记录

本次生产部署已在服务器 `43.160.242.244` 完成。

### 服务器架构

- 服务器系统/架构：`linux/amd64`
- 本地构建机器架构：`arm64`
- 最终部署镜像：`new-api:prod-20260605-b0ac0429-amd64`
- 运行容器：`new-api`
- 公网入口：`http://43.160.242.244:3000`

第一次尝试是在 Apple Silicon 本地直接构建完整 `linux/amd64` Docker 镜像，但前端构建阶段失败。失败原因是在 QEMU 下运行 Bun/Rspack 时，加载原生模块 `@rspack/binding-linux-x64-gnu` 触发 `SIGILL`。

最终采用的可行构建路径：

1. 使用 arm64 Bun 容器在本地构建前端静态资源。
2. 在本地交叉编译 `linux/amd64` Go 二进制。
3. 只打包运行二进制、许可证文件和部署文件。
4. 上传压缩运行包到服务器。
5. 在 amd64 服务器上构建最终运行镜像。

服务器上没有保留源码目录。`/opt/services/new-api/src` 已删除并确认不存在。

### 构建产物

- 本地压缩运行包：`/tmp/new-api-server-runtime-b0ac0429.tgz`
- 服务器运行包路径：`/opt/services/new-api/new-api-server-runtime-b0ac0429.tgz`
- 运行包 SHA-256：

```text
0b6d91e11fa7d53cf279b825e42e362b19acbb13f38b17ec2a4fdcd55f1b5359
```

- 交叉编译二进制 SHA-256：

```text
1f2f038cc48e3f7cd3d71781715d7fc6b1c9e75e71a9ff0c10b23a143e91abae
```

服务器验证运行二进制为：

```text
ELF 64-bit LSB executable, x86-64, statically linked
```

服务器验证 Docker 镜像为：

```text
new-api:prod-20260605-b0ac0429-amd64 amd64 linux
```

### 实际运行时路由

初始 compose 配置使用了 Docker 服务名 `sub2api-postgres` 和 `sub2api-redis`。这在本服务器上不可用，因为这些服务通过 host 网络访问，而不是从 `new-api` 容器内通过 Docker DNS 解析。

已部署配置使用 Docker host-gateway：

```yaml
extra_hosts:
  - "host.docker.internal:host-gateway"
```

生产 `.env` 使用：

```env
SQL_DSN=postgresql://new_api:<redacted>@host.docker.internal:5432/new_api?sslmode=disable
REDIS_CONN_STRING=redis://:<redacted>@host.docker.internal:6379/1
```

真实密钥只保存在服务器 `/opt/services/new-api/.env`。

### 数据库与 Redis 隔离

- 已创建 PostgreSQL 角色 `new_api`。
- 已创建 PostgreSQL 数据库 `new_api`。
- `new-api` 迁移只在 `new_api` 中执行，没有连接 `sub2api` 数据库。
- 部署后 `new_api` 的 public schema 中有 `26` 张表。
- 现有 `sub2api` 数据库仍然独立存在。
- Redis 复用现有 Redis 服务，但使用 DB `1`。
- Redis DB `0` 继续保留给 `sub2api` 使用。

### 验证结果

部署验证已通过：

```text
new-api container: Up / healthy
image architecture: amd64 linux
local status:  http://127.0.0.1:3000/api/status success=true
public status: http://43.160.242.244:3000/api/status success=true
sub2api: healthy
sub2api-postgres: running
sub2api-redis: running
```

已部署的 `new-api` 日志显示：

```text
using PostgreSQL as database
database migration started
Redis is enabled
New API started
```

### 运维备注

- 服务器通过普通 `scp` 传输文件不稳定，最终使用 `rsync -avP --partial --inplace` 完成断点续传。
- 已为本地机器启用 SSH key 登录，避免部署过程中反复输入密码。
- 运行包保留在服务器上，便于审计和重建镜像。
- 不要删除 `/opt/services/new-api/.env`，其中保存了生产数据库、Redis 和会话密钥。
