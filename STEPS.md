# PassNow — Quick Local Setup

## 1. Pull latest code

```powershell
git pull origin main
```

## 2. Configure `.env`

Set your platform database and local URLs:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=passnow_platform
DB_USER=root
DB_PASSWORD=

API_BASE_URL=http://192.168.100.11:8080
APP_BASE_URL=http://192.168.100.11:8000
BASE_DOMAIN=passnow.test
```

For tenant database provisioning, also set:

TENANT_DB_ENCRYPTION_KEY=your-32-byte-base64-key

## 3. Run Platform migrations

go run ./cmd/migrate -scope platform -action up

Check status:

go run ./cmd/migrate -scope platform -action status

## 4. Tenant migrations

Tenant migrations normally run automatically when a tenant is created from Platform Admin.

Manual test:

```powershell
go run ./cmd/migrate -scope tenant -action up -database passnow_glee_hotel -user root
```

Check status:

```powershell
go run ./cmd/migrate -scope tenant -action status -database albatech_passnow -user alumasinde -password "21082108"
```

## 5. Test backend

```powershell
go test ./...
```

## 6. Start Go API

```powershell
go run ./cmd/api
```

API:

```text
http://192.168.100.11:8080
```

## 7. Start PHP frontend

```powershell
php -S 0.0.0.0:8000 -t frontend/public frontend/public/router.php
```

Frontend:

```text
http://192.168.100.11:8000/login
```
