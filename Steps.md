# PassNow — Quick Local Setup

## 1. Pull latest code

```powershell
git pull origin main
```

## 2. Configure `.env`

Set your database values and:

```env
API_BASE_URL=http://192.168.100.11:8080
APP_BASE_URL=http://192.168.100.11:8000
```

## 3. Run migrations

```powershell
go run ./cmd/migrate -action up
```

## 4. Test backend

```powershell
go test ./...
```

## 5. Start Go API

```powershell
go run ./cmd/api
```

## 6. Start frontend

From the repository root:

```powershell
php -S 0.0.0.0:8000 -t frontend/public frontend/public/router.php
```

Then open:

```text
http://192.168.100.11:8000/login
```

Do not use `php -S 0.0.0.0:8000 frontend/public/router.php` without `-t frontend/public`; PHP can resolve the relative router path incorrectly for incoming requests.

