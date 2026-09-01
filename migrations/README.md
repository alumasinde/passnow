# PassNow Migrations

PassNow uses two independent migration scopes.

## Platform

Directory:

`migrations/platform`

Run against the single PassNow platform database. It contains only platform-owned data such as platform users/admins, tenants, tenant database metadata, and tenant domains.

```powershell
go run ./cmd/migrate -scope platform -action up
go run ./cmd/migrate -scope platform -action status
```

## Tenant

Directory:

`migrations/tenant`

Run against one tenant database. It contains tenant application tables such as users, roles, employees, visitors, visits, gatepasses, approvals, settings, and audit data.

```powershell
go run ./cmd/migrate -scope tenant -action up -database passnow_glee_hotel -user root
go run ./cmd/migrate -scope tenant -action status -database passnow_glee_hotel -user root
```

Optional connection overrides:

```powershell
go run ./cmd/migrate -scope tenant -action up -host 127.0.0.1 -port 3306 -database passnow_glee_hotel -user root -password "your_password"
```

The root `migrations/` directory is legacy and is no longer an executable migration scope. New migrations must be added only to `migrations/platform/` or `migrations/tenant/`.

Tenant provisioning automatically runs the tenant migration set when a tenant is created.
