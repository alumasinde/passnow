# PassNow Frontend — Batch 5

## Added
- Reusable ResourcePage API-backed listing helper
- Visitors list/search/filter/pagination
- Visitor creation and details
- Visitor blacklist operation
- Visits list/search/filter/pagination
- Visit creation and details
- Visit check-in/check-out
- Employees list/search/filter/pagination
- Employee creation and details
- Reusable forms and existing layout/components
- CSRF-protected mutations
- Explicit API route mapping
- Tenant/business authorization remains backend-owned

## Design boundary
The PHP frontend is a presentation/BFF layer. It validates basic input, protects browser mutations with CSRF, and calls known Go API routes. It does not become the source of truth for tenant configuration, permissions, workflow states, or security rules.
