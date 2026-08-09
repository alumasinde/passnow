# PassNow Frontend — Batch 6

## Administration and tenant configuration

Added:
- Settings hub
- Users
- User details
- User invitations
- Roles
- Role permissions
- Approval workflows
- Approval workflow editor
- Gatepass types
- Visit types
- ID types
- Visitor companies
- Departments
- Gatepass settings
- Visitor settings
- Reusable AdminResource and ConfigCrud helpers

## Security/design
- All protected pages require authentication.
- Browser mutations require CSRF validation.
- API routes are explicit rather than user-supplied.
- Tenant authorization and supported configuration remain enforced by the Go backend.
- Frontend does not hardcode tenant role names, workflow names, types, departments or configuration values.
