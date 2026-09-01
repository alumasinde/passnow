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


## Platform Admin (Phase 1)

After creating or bootstrapping a user, grant that existing user PassNow platform access:

    go run ./cmd/platform-admin -email admin@example.com -role owner

Platform login endpoint:

    POST /api/v1/platform/auth/login

The returned platform token is only valid for platform routes and cannot be used against tenant APIs.
