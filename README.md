# PassNow

PassNow is a multi-tenant visitor, visit, approval and gatepass platform built with Go and MySQL.

## Requirements

- Go 1.22+
- MySQL 8+
- Docker + Docker Compose (optional)

## Local setup

Copy the environment template and set real values:

```bash
cp env.example .env
```

Never commit `.env` or production secrets.

### Verify the code without MySQL

```bash
go build ./...
go test ./...
```

### Run with MySQL

```bash
go run ./cmd/passnow migrate

go run ./cmd/passnow migrate status
go run ./cmd/passnow serve
```

Health endpoints:

- `GET /healthz` — process is alive.
- `GET /readyz` — application can reach MySQL.

### Docker development environment

This starts an isolated MySQL 8.4 instance, applies migrations, and starts PassNow:

```bash
docker compose -f docker-compose.dev.yml up --build
```

Or:

```bash
make docker-dev
```

The credentials in `docker-compose.dev.yml` are for local development only. Replace them for any shared or production environment.

## Migration commands

```bash
go run ./cmd/passnow migrate
go run ./cmd/passnow migrate status
```

The migration connection is separate from the application connection. MySQL `multiStatements` is enabled only for migrations.

## Architecture

```text
HTTP
  -> middleware
  -> handlers
  -> services
  -> repositories
  -> MySQL
```

Tenants share one database schema and are isolated by `tenant_id` at the repository/domain layer. Tenant-specific configuration belongs in tenant data, not separate environment files.

## Production notes

- Use a strong randomly generated `JWT_SECRET` of at least 32 characters.
- Keep `PLATFORM_BOOTSTRAP_TOKEN` empty after initial provisioning or rotate/remove it according to the deployment process.
- Put TLS termination in the reverse proxy/load balancer.
- Do not use the development Docker Compose passwords in production.
- Run migrations as a controlled deployment step before starting application instances.
- Monitor `/readyz` for database readiness and `/healthz` for process liveness.
