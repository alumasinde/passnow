# Batch 5 — Security & multi-tenancy hardening

## Tenant isolation

Every authenticated request now requires:
1. resolved tenant from host/path
2. JWT tenant claim matching resolved tenant
3. active tenant membership
4. membership role matching the JWT role claim
5. current permission lookup for that role

This prevents a stale JWT role from retaining privileges after a role change.

## Refresh-token binding

Refresh tokens now carry `tenant_id`. A token issued in tenant A cannot be
used at tenant B's refresh endpoint, even when the same user belongs to both.

Legacy refresh tokens that belong to users in multiple tenants are revoked
during migration because there is no safe tenant to infer.

## Host normalization

Host/domain matching is lower-cased and trailing-dot normalized. Platform
domain matching requires either an exact match or a `.<base-domain>` suffix.

## Rate limiting

Login and refresh endpoints have process-local rate limits. Do not treat this
as the only production control when running multiple API replicas: put a
shared gateway/WAF rate limiter in front of the API as well.

The limiter intentionally uses RemoteAddr and does not trust arbitrary
X-Forwarded-For headers.

## QR

QR images continue to use `Cache-Control: no-store`. QR lookup remains tenant
scoped and returns NotFound for a token belonging to another tenant.

## CSRF

The API uses Authorization bearer tokens rather than browser cookies for the
authenticated API, so CSRF tokens are not required for these API endpoints.
If a browser cookie-based session is introduced later, add CSRF protection.

## Database

Apply migration `0012_security_hardening.up.sql` before deploying the new
refresh-token code.
