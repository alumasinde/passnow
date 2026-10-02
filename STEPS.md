# PassNow — Local Development Setup

This document explains how to run PassNow locally with:

* **Platform:** `passnow.test`
* **Tenant:** `<tenant>.passnow.test`
* **PHP Frontend:** port `8000`
* **Go API:** port `8080`
* **MySQL/MariaDB:** local database
* **Redis:** local Redis instance

---

## 1. PassNow URL Architecture

PassNow uses the hostname to determine the tenant.

### Platform

```text
http://passnow.test:8000
```

Platform login:

```text
http://passnow.test:8000/platform/login
```

Platform dashboard:

```text
http://passnow.test:8000/platform/tenants
```

### Tenant

A tenant uses its own subdomain:

```text
http://<tenant-slug>.passnow.test:8000
```

For example, if the tenant slug is `ghl`:

```text
http://ghl.passnow.test:8000
```

Tenant login:

```text
http://ghl.passnow.test:8000/login
```

Tenant dashboard:

```text
http://ghl.passnow.test:8000/dashboard
```

Another tenant:

```text
http://abc.passnow.test:8000/login
```

The tenant is determined from the hostname.

---

# 2. Required Software

Install the following:

* PHP 8.x or the version required by the project
* Composer
* Go
* MySQL/MariaDB
* Redis
* Git

Verify:

```bash
php -v
composer -V
go version
mysql --version
redis-cli --version
```

---

# 3. Project Structure

The project contains two main applications:

```text
passnow/
├── api/
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   └── ...
│
├── frontend/
│   ├── app/
│   ├── public/
│   └── ...
│
└── ...
```

The frontend runs on:

```text
http://localhost:8000
```

The API runs on:

```text
http://localhost:8080
```

---

# 4. Configure Local Hostnames

Because PassNow uses subdomains, Windows must resolve the domains to your local machine.

Open Notepad as Administrator.

Open:

```text
C:\Windows\System32\drivers\etc\hosts
```

Add:

```text
127.0.0.1 passnow.test
127.0.0.1 ghl.passnow.test
127.0.0.1 abc.passnow.test
127.0.0.1 xyz.passnow.test
```

Add additional tenant domains when needed:

```text
127.0.0.1 <tenant-slug>.passnow.test
```

Save the file.

Test:

```powershell
ping passnow.test
ping ghl.passnow.test
```

Both should resolve to:

```text
127.0.0.1
```

---

# 5. Configure the Environment

Create the appropriate `.env` file.

Use:

```env
APP_ENV=local

APP_BASE_URL=

BASE_DOMAIN=passnow.test

API_BASE_URL=http://127.0.0.1:8080
```

## Important

Do NOT use:

```env
APP_BASE_URL=http://192.168.100.11:8000
```

for the hostname-based tenant architecture.

Do NOT use:

```env
APP_BASE_URL=http://passnow.test:8000
```

for normal application URL generation either.

`APP_BASE_URL` should remain empty if the application generates relative URLs.

This prevents:

```text
ghl.passnow.test
        ↓
redirect
        ↓
192.168.100.11
```

and keeps the browser on the current tenant hostname.

---

# 6. Tenant Hostname Resolution

PassNow determines the tenant from:

```text
HTTP_HOST
```

Example:

```text
ghl.passnow.test:8000
```

The application should extract:

```text
ghl
```

as the tenant slug.

For:

```text
abc.passnow.test:8000
```

the tenant slug is:

```text
abc
```

The port must be ignored when resolving the tenant.

Conceptually:

```text
HTTP_HOST
    │
    ▼
ghl.passnow.test:8000
    │
    ▼
remove port
    │
    ▼
ghl.passnow.test
    │
    ▼
remove .passnow.test
    │
    ▼
ghl
    │
    ▼
tenant lookup
```

---

# 7. Platform Hostname

The platform hostname is:

```text
passnow.test
```

It is NOT a tenant.

Therefore:

```text
passnow.test
```

must be treated as the platform host.

Tenant hosts are:

```text
ghl.passnow.test
abc.passnow.test
xyz.passnow.test
```

The platform must never attempt to resolve:

```text
passnow
```

as a tenant.

---

# 8. Start Redis

Start Redis.

Verify:

```bash
redis-cli ping
```

Expected:

```text
PONG
```

---

# 9. Create the Database

Create the main PassNow database.

Example:

```sql
CREATE DATABASE passnow;
```

Configure the database credentials in `.env`.

Example:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=passnow
DB_USERNAME=root
DB_PASSWORD=
```

Use the actual credentials for your local machine.

---

# 10. Run API Migrations

From the API directory:

```bash
go run ./cmd/migrate up
```

---

# 11. Install PHP Dependencies

From the frontend directory:

```bash
composer install
```

If dependencies are already installed:

```bash
composer dump-autoload
```

---

# 12. Start the Go API

From the API directory:

```bash
go run ./cmd/api
```

The API should listen on:

```text
127.0.0.1:8080
```

Test:

```text
http://127.0.0.1:8080
```

or use the project's health endpoint if one exists.

---

# 13. Start the PHP Frontend

From the frontend directory:

```bash
php -S 0.0.0.0:8000 -t public
```

The application is now available on:

```text
http://localhost:8000
```

But for PassNow tenant testing, use the configured hostnames.

---

# 14. Access Platform

Open:

```text
http://passnow.test:8000/platform/login
```

Platform login should authenticate against the platform authentication system.

After successful login, the platform should redirect to:

```text
http://passnow.test:8000/platform/tenants
```

The platform session must remain on:

```text
passnow.test
```

---

# 15. Access a Tenant

If the tenant slug is:

```text
ghl
```

open:

```text
http://ghl.passnow.test:8000/login
```

After login:

```text
http://ghl.passnow.test:8000/dashboard
```

The browser must remain on:

```text
ghl.passnow.test
```

It must NOT redirect to:

```text
192.168.100.11
```

and must NOT redirect to:

```text
passnow.test
```

during normal tenant navigation.

---

# 16. Tenant API Requests

When the tenant frontend communicates with the API, it must preserve the current tenant host.

For:

```text
ghl.passnow.test:8000
```

the frontend should send the tenant context to the API.

The API should receive:

```text
Host: ghl.passnow.test
```

or the equivalent tenant context mechanism implemented by the project.

The API then resolves:

```text
ghl
```

and loads the correct tenant.

---

# 17. Login Flow

### Tenant login

```text
Browser
   │
   ▼
ghl.passnow.test:8000/login
   │
   ▼
Tenant hostname detected
   │
   ▼
Tenant = ghl
   │
   ▼
Login submitted
   │
   ▼
Go API
   │
   ▼
Tenant context = ghl
   │
   ▼
Authentication
   │
   ▼
Session/token created
   │
   ▼
/dashboard
```

The final URL should be:

```text
http://ghl.passnow.test:8000/dashboard
```

---

# 18. Platform Login Flow

```text
Browser
   │
   ▼
passnow.test:8000/platform/login
   │
   ▼
Platform authentication
   │
   ▼
Platform session
   │
   ▼
/platform/tenants
```

Final URL:

```text
http://passnow.test:8000/platform/tenants
```

---

# 19. Important URL Rule

Application redirects should use relative URLs.

Correct:

```text
/login
/dashboard
/platform/login
/platform/tenants
```

Avoid hard-coded URLs such as:

```text
http://192.168.100.11:8000/login
```

or:

```text
http://passnow.test:8000/login
```

inside tenant navigation.

The browser should preserve the current hostname.

For example:

```text
Current:
http://ghl.passnow.test:8000/login

redirect('/dashboard')

Result:
http://ghl.passnow.test:8000/dashboard
```

---

# 20. Do Not Hard-Code Tenant Slugs

Do not write:

```php
$tenant = 'ghl';
```

Tenant identity must come from:

```text
HTTP_HOST
```

Therefore the same application supports:

```text
ghl.passnow.test
abc.passnow.test
xyz.passnow.test
```

without code changes.

---

# 21. Testing Multiple Tenants

Create two tenants:

```text
ghl
abc
```

Then test:

### Tenant 1

```text
http://ghl.passnow.test:8000/login
```

Login and verify:

```text
http://ghl.passnow.test:8000/dashboard
```

### Tenant 2

```text
http://abc.passnow.test:8000/login
```

Login and verify:

```text
http://abc.passnow.test:8000/dashboard
```

The two tenants must not share tenant data.

---

# 22. Test Platform

Open:

```text
http://passnow.test:8000/platform/login
```

Login.

Verify:

```text
http://passnow.test:8000/platform/tenants
```

Then open:

```text
http://ghl.passnow.test:8000/login
```

Verify that the tenant login remains on:

```text
ghl.passnow.test
```

---

# 23. Common Problem: Redirects to IP Address

If:

```text
http://ghl.passnow.test:8000/login
```

redirects to:

```text
http://192.168.100.11:8000/login
```

check `.env`.

Incorrect:

```env
APP_BASE_URL=http://192.168.100.11:8000
```

Correct:

```env
APP_BASE_URL=
```

Also search the project for:

```text
192.168.100.11
```

and remove hard-coded frontend redirect URLs where they are not required.

---

# 24. Common Problem: Tenant Cannot Login

Check:

### 1. Hostname

```text
ghl.passnow.test
```

must resolve locally.

### 2. BASE_DOMAIN

```env
BASE_DOMAIN=passnow.test
```

### 3. Tenant exists

The database must contain:

```text
slug = ghl
```

### 4. API is running

```text
127.0.0.1:8080
```

### 5. Browser remains on tenant host

After login, it should remain:

```text
ghl.passnow.test
```

not:

```text
192.168.100.11
```

---

# 25. Clear Browser Session When Testing

When changing authentication or hostname configuration, clear existing cookies for:

```text
passnow.test
```

and:

```text
ghl.passnow.test
```

Then restart the PHP and Go applications.

Test again from:

```text
http://ghl.passnow.test:8000/login
```

---

# 26. Production Architecture

The same architecture should eventually become:

```text
passnow.com
    │
    └── Platform

ghl.passnow.com
    │
    └── GHL tenant

abc.passnow.com
    │
    └── ABC tenant
```

The local development equivalent is:

```text
passnow.test
    │
    └── Platform

ghl.passnow.test
    │
    └── GHL tenant

abc.passnow.test
    │
    └── ABC tenant
```

---

# 27. Final URL Reference

| Purpose          | URL                                         |
| ---------------- | ------------------------------------------- |
| Platform         | `http://passnow.test:8000`                  |
| Platform Login   | `http://passnow.test:8000/platform/login`   |
| Platform Tenants | `http://passnow.test:8000/platform/tenants` |
| GHL Tenant       | `http://ghl.passnow.test:8000`              |
| GHL Login        | `http://ghl.passnow.test:8000/login`        |
| GHL Dashboard    | `http://ghl.passnow.test:8000/dashboard`    |
| ABC Tenant       | `http://abc.passnow.test:8000`              |
| ABC Login        | `http://abc.passnow.test:8000/login`        |
| API              | `http://127.0.0.1:8080`                     |

---

# 28. Golden Rule

**Platform is identified by the root domain. Tenants are identified by subdomains.**

```text
passnow.test
     │
     └── PLATFORM

ghl.passnow.test
     │
     └── TENANT: ghl

abc.passnow.test
     │
     └── TENANT: abc
```

Never replace a tenant hostname with the server IP during normal application navigation.

Never hard-code a tenant slug.

Never use `APP_BASE_URL` to force all tenant URLs onto one hostname.

The browser's current hostname is part of the tenant context and must be preserved.
