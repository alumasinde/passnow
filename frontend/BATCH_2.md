# PassNow Frontend — Batch 2

## Added

- Reusable `ListQuery`
- Reusable `Paginator`
- Reusable data table
- Search toolbar
- Dynamic filters
- Rows-per-page selector
- Pagination with ellipsis
- Empty states
- Reusable modal component
- Modal JS controller
- Responsive table foundation
- Gatepasses listing page wired to `/api/v1/gatepasses`

## Design rule

The frontend does not define business status/type lists as fixed PHP constants. When the API supplies metadata (`meta.statuses`, `meta.types`) those values are used. Query parameters are passed through only for known filters.

The list UI is reusable for Visitors, Visits, Employees, Approvals, Audit Logs and Reports in later batches.

## API shape

The gatepass page defensively supports:

```json
{
  "data": [],
  "meta": {
    "total": 0,
    "statuses": [],
    "types": []
  }
}
```

and common alternatives such as `items`, `results`, and top-level `total`.

The exact Go API response remains the source of truth; later batches can tighten the mapping once endpoint contracts are verified.
