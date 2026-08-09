# PassNow Frontend — Batch 3

Gatepass workflow UI.

## Added

- Dynamic gatepass creation form
- Tenant-provided gatepass type options
- Gatepass detail page
- State-aware approval/check-out/check-in actions
- Movement history
- Reusable modal confirmations
- QR display through the backend QR endpoint
- Returnable gatepass UX
- Reusable form field/select/textarea components
- Server-side validation before API calls
- CSRF protection on all mutating frontend forms
- Explicit operation-to-API route mapping (no arbitrary route input)
- No business authorization decisions in PHP; Go remains authoritative

## Important

The frontend deliberately does not invent gatepass statuses or approval rules. It only uses states returned by the API to improve UX. The Go backend must enforce every transition, permission, tenant scope, returnable rule, and approval requirement.
