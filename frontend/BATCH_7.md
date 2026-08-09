# PassNow Frontend — Batch 7

## Shared product-quality layer

Added:
- Reusable flash/toast notifications
- Confirmation modal system
- Loading/spinner states for forms
- Global interaction JavaScript
- Reusable CSV export for visible tables
- Shared 403 / 404 / 500 pages
- Shared UI component injection into layouts
- Responsive interaction styling
- Reusable admin resource/config helpers retained
- Export buttons on major list screens where the existing table component is available

## Security boundary

The frontend remains a presentation/BFF layer:
- CSRF is required for browser mutations.
- Authentication is required on protected pages.
- Business authorization remains in Go.
- Tenant isolation remains in Go.
- CSV export exports the currently rendered table only; it does not bypass API permissions or fetch hidden records.

## Important

For production, web-server routing should map unknown routes to the 404 page and unexpected application exceptions to the 500 page. The Go API remains the authoritative source for permissions, filtering, pagination and tenant-scoped data.
