# PassNow — Quick Local Setup

## 1. Pull latest changes

```powershell
git pull origin main
```

## 2. Configure `.env`

Ensure your database configuration is set and use your laptop IP addresses:

```env
API_BASE_URL=http://192.168.100.11:8080
APP_BASE_URL=http://192.168.100.11:8000
```

## 3. Tenant database encryption (Phase A)

Add a 32-byte base64 key to `.env` before using tenant database provisioning:

```powershell
$bytes = New-Object byte[] 32; [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes); [Convert]::ToBase64String($bytes)
```

Then add the generated value:

```env
TENANT_DB_ENCRYPTION_KEY=generated-value
TENANT_DB_MAX_OPEN_CONNS=10
TENANT_DB_MAX_IDLE_CONNS=5
TENANT_DB_CONN_MAX_LIFETIME=5m
```

## 4. Run database migrations

```powershell
go run ./cmd/migrate -action up
```

Migrations that have already been applied will be skipped.

## 5. Test the backend

```powershell
go test ./...
```

## 6. Start the Go API

```powershell
go run ./cmd/api
```

API:

```text
http://192.168.100.11:8080
```

## 7. Start the PHP frontend

From the project root:

```powershell
php -S 0.0.0.0:8000 -t frontend/public frontend/public/router.php
```

Frontend:

```text
http://192.168.100.11:8000/login
```
