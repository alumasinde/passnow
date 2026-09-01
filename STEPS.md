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

## 3. Run database migrations

```powershell
go run ./cmd/migrate -action up
```

Migrations that have already been applied will be skipped.

## 4. Test the backend

```powershell
go test ./...
```

## 5. Start the Go API

```powershell
go run ./cmd/api
```

API:

```text
http://192.168.100.11:8080
```

## 6. Start the PHP frontend

From the project root:

```powershell
php -S 0.0.0.0:8000 -t frontend/public frontend/public/router.php
```

Frontend:

```text
http://192.168.100.11:8000/login
```
