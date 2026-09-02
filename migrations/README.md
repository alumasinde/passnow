# PassNow Migrations

PassNow has exactly two executable migration scopes:

```text
migrations/
├── platform/   # PassNow platform database only
└── tenant/     # One isolated tenant application database
```

Do not add executable `.up.sql` files directly under `migrations/`.

## Platform migrations

Run against the single PassNow platform database.

Platform tables include:

- `tenants`
- `users` used for platform administration
- `platform_admins`
- `platform_audit_logs`
- `tenant_databases`
- `tenant_domains`

Run:

```powershell
go run ./cmd/migrate -scope platform -action up
go run ./cmd/migrate -scope platform -action status
```

## Tenant migrations

Run against one tenant's own database.

Tenant tables include:

- users and roles
- permissions and memberships
- audit logs
- settings
- departments and employees
- ID types and visitor companies
- visitors and visits
- approval workflows
- gatepass types, gatepasses, approvals and items
- gatepass movements
- notification outbox

Tenant migrations must never create:

- a `tenants` table
- a `tenant_id` column for row isolation
- foreign keys to a platform `tenants` table

Tenant isolation is provided by the database connection itself.

Run manually:

```powershell
go run ./cmd/migrate -scope tenant -action up -database passnow_glee_hotel -user root
go run ./cmd/migrate -scope tenant -action status -database passnow_glee_hotel -user root
```

Recommended for a provisioned tenant (loads and decrypts the stored credentials):

```powershell
go run ./cmd/migrate -scope tenant -action status -tenant-id 1
```

With explicit connection values:

```powershell
go run ./cmd/migrate -scope tenant -action up -host 127.0.0.1 -port 3306 -database passnow_glee_hotel -user root -password "your_password"
```

Tenant provisioning automatically runs `migrations/tenant/` after the tenant database is created or verified.

## Adding a new migration

Platform-only change:

```text
migrations/platform/0004_example.up.sql
```

Tenant application change:

```text
migrations/tenant/0014_example.up.sql
```

Never edit a migration that has already been applied to a database. Add a new migration instead.
