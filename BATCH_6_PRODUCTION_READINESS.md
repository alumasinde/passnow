# Batch 6 — Integration & production readiness

## Lifecycle

### Approval-required
draft/request -> pending_approval -> approved -> physical checkout ->
awaiting_return -> partial returns -> completed.

Rejection/cancellation/expiry terminate the authorization path.

### No approval required
Creation authorizes the pass immediately:
created -> approved -> physical checkout -> completed/awaiting_return.

The creation transaction now records `issued_by` and `issued_at` for this path,
so the expiry worker can correctly expire an approved pass that was never used.

## Physical movement

The movement tables are the authoritative physical audit trail:
- gatepass_movements
- gatepass_movement_items

The header timestamps remain summary/convenience fields.

A final full return moves the pass to `completed`; partial returns remain
`partially_returned` until all outstanding quantities are accounted for.

## Database tenant isolation

Migration 0013 adds composite tenant-bound foreign keys for critical relationships.
This protects against cross-tenant references even if application code later
makes a tenant-filtering mistake.

Before production, run the migration against a copy/staging database first.
If historical data contains cross-tenant references, correct those rows before
applying the constraints.

## Deployment checklist

1. Apply migrations in order through 0013.
2. Run `go mod tidy`.
3. Run `go test ./...`.
4. Run `go vet ./...`.
5. Use a dedicated MySQL application user with no schema-altering privileges.
6. Put TLS termination and a shared rate limiter/WAF in front of multiple API replicas.
7. Set a strong random JWT_SECRET (>= 32 characters; preferably 256-bit random).
8. Set APP_ENV=production.
9. Back up MySQL before migrations.
10. Test tenant A/B isolation with real staging accounts.
