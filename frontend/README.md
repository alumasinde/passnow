# PassNow PHP Frontend — Batch 1

Server-rendered PHP frontend for the Go API in this repository.

## Security model
- Go remains the business/data layer.
- Browser never receives the Go refresh token.
- Access and refresh tokens are kept in the PHP server session.
- CSRF tokens protect PHP state-changing forms.
- Session cookies are HttpOnly and configurable for Secure/SameSite.
- API calls use cURL with TLS verification.
- Tenant identity is resolved by the Go API from the tenant host; PHP forwards the configured tenant host to the API.

## Environment
APP_NAME=PassNow
APP_BASE_URL=http://localhost:8000
API_BASE_URL=http://127.0.0.1:8080
TENANT_HOST=tenant.example.com
APP_TIMEZONE=Africa/Nairobi
SESSION_NAME=passnow_session
SESSION_SECURE=0
SESSION_SAMESITE=Lax
CSRF_TTL=3600
API_TIMEOUT=15
PAGE_SIZE=20
FONT_AWESOME_URL=https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css

For production, use HTTPS and set SESSION_SECURE=1.

## Run
From the repository root:

php -S 127.0.0.1:8000 -t frontend/public

The Go API must be running at API_BASE_URL.

## Batch 1
Foundation only: application bootstrap, secure API client, authentication/session, refresh handling, CSRF, reusable layouts/partials/components, dashboard API integration, responsive CSS and vanilla JS.
