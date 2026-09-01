## Backend Setup

Configure your `.env` with your local MySQL and application settings. Make sure MySQL is running.

### Run migrations

go run ./cmd/migrate -action up

### Run tests

go test ./...

### Start the backend

go run ./cmd/api

Backend health check:

http://localhost:8080/healthz

---

## Frontend

Start Apache and MySQL from XAMPP.

Open the frontend using the configured local URL.

The frontend communicates with the Go backend API.

---

## Testing from Phone

Ensure your laptop and phone are connected to the same Wi-Fi.

Find your laptop IP:

ipconfig

Example:

text
192.168.100.11


Test the backend:

text
http://192.168.100.11:8080/healthz


For tenant API routes:

text
http://192.168.100.11:8080/TENANT-SLUG/api/v1/...


Make sure the backend listens on:

text
:8080


and Windows Firewall allows port `8080`.

---

## Daily Development Workflow


git pull origin main

go run ./cmd/migrate -action up

go test ./...

go run ./cmd/api


Before pushing changes:


go test ./...
git status
git add .
git commit -m "describe your change"
git push origin main

