# PassNow Quick Setup

## 1. Create `.env`

Copy `.env.example` to `.env` and update your MySQL credentials, JWT secret, and `TENANT_HOST`.

## 2. Backend

From the repository root:

    go run ./cmd/migrate -action up
    go test ./...
    go run ./cmd/api

Health check:

    http://localhost:8080/healthz

## 3. Frontend

Open another terminal from the repository root:

    php -S 127.0.0.1:8000 -t frontend/public

Open:

    http://localhost:8000

The frontend and backend both read the same root `.env`.

## Daily workflow

    git pull origin main
    go run ./cmd/migrate -action up
    go test ./...
    go run ./cmd/api

In another terminal:

    php -S 127.0.0.1:8000 -t frontend/public
