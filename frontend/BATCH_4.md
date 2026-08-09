# PassNow Frontend — Batch 4

## Added

### Approval Center
- Pending approvals list
- Search/filter/pagination
- Approval detail page
- Approval chain/timeline
- Approve/reject with comment
- CSRF protection
- Explicit backend route mapping

### Gate Operations
- QR token lookup
- Gate verification screen
- Physical verification confirmation
- Check-out
- Check-in
- Returnable-item warning on check-in
- Movement actions linked to the existing Go API

## Security boundary

PHP only controls presentation and request validation. The Go API remains authoritative for:
- tenant scope
- permissions
- approval eligibility
- approval order
- state transitions
- returnable enforcement
- check-in/check-out authorization
- audit/movement integrity

The frontend never trusts a status or permission merely because it was rendered. Mutating requests are sent to the protected backend, which must re-check authorization.
