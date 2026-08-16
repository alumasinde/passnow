# PassNow PHP Frontend

Server-rendered PHP frontend/BFF for the Go API in this repository.

## Security model
- Go remains the business/data and authorization layer.
- Browser never receives the Go refresh token.
- Access and refresh tokens are kept in the PHP server session.
- CSRF tokens protect PHP state-changing forms and rotate after authentication.
- Session cookies are HttpOnly, SameSite-controlled and Secure in production.
- API calls use cURL with TLS verification and do not follow redirects.
- API requests use relative paths only and include a request ID for correlation.
- Tenant identity is resolved authoritatively by the Go API from the tenant host; PHP only forwards the validated host value.
- Frontend authorization is presentation-only; Go must enforce every permission and tenant boundary.

## Environment
```text
APP_NAME=PassNow
APP_ENV=development
APP_BASE_URL=http://localhost:8000
API_BASE_URL=http://127.0.0.1:8080
TENANT_HOST=tenant.example.com
APP_TIMEZONE=Africa/Nairobi
SESSION_NAME=passnow_session
SESSION_SECURE=0
SESSION_SAMESITE=Lax
SESSION_IDLE_TIMEOUT=1800
SESSION_ABSOLUTE_TIMEOUT=28800
CSRF_TTL=3600
API_TIMEOUT=15
PAGE_SIZE=20
FONT_AWESOME_URL=https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css
```

For production:
- Use HTTPS for both the frontend and API where applicable.
- Set `APP_ENV=production`.
- Set `SESSION_SECURE=1`.
- Use the actual tenant host/domain expected by the Go API.
- Keep `API_BASE_URL` private/reachable only from the frontend server where possible.
- Never put access or refresh tokens in browser JavaScript, HTML, URLs or logs.

## Run
From the repository root:

```bash
php -S 127.0.0.1:8000 -t frontend/public
```

The Go API must be running at `API_BASE_URL`.

## Production hardening included
- Strict PHP sessions with cookies only and no URL-based session IDs.
- Idle and absolute session expiration.
- CSRF expiration and rotation.
- Safe internal redirects; external redirects require explicit `redirectExternal()` use.
- API response handling for JSON and `204 No Content` responses.
- API transport errors do not expose cURL internals to users.
- Validated tenant host before sending the `Host` header to the Go API.
- Shared 419 security-token error page.
