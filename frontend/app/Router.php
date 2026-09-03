<?php
declare(strict_types=1);

final class Router
{
    /** @var array<string,string> */
    private array $routes = [
        '/' => 'Auth/Login.php',
        '/login' => 'Auth/Login.php',
        '/logout' => 'Auth/Logout.php',
        '/dashboard' => 'Dashboard/Index.php',
        '/visitors' => 'Visitors/Index.php','/visitors.php'=>'Visitors/Index.php','/visitors/view'=>'Visitors/Show.php','/visitor.php'=>'Visitors/Show.php','/visitors/create'=>'Visitors/Create.php','/visitor-create.php'=>'Visitors/Create.php','/visitors/edit'=>'Visitors/Edit.php','/visitor-edit.php'=>'Visitors/Edit.php','/visitors/blacklist'=>'Visitors/Blacklist.php','/visitor-blacklist.php'=>'Visitors/Blacklist.php','/visitors/companies'=>'Visitors/Companies.php','/visitor-companies.php'=>'Visitors/Companies.php','/visitors/companies/edit'=>'Visitors/CompanyEdit.php','/visitor-companies-edit.php'=>'Visitors/CompanyEdit.php','/visitors/settings'=>'Visitors/Settings.php','/visitor-settings.php'=>'Visitors/Settings.php',
        '/visits'=>'Visits/Index.php','/visits.php'=>'Visits/Index.php','/visits/view'=>'Visits/Show.php','/visit.php'=>'Visits/Show.php','/visits/create'=>'Visits/Create.php','/visit-create.php'=>'Visits/Create.php','/visits/operation'=>'Visits/Operation.php','/visit-operation.php'=>'Visits/Operation.php','/visits/types'=>'Visits/Types.php','/visit-types.php'=>'Visits/Types.php','/visits/types/edit'=>'Visits/TypeEdit.php','/visit-types-edit.php'=>'Visits/TypeEdit.php',
        '/gatepasses'=>'Gatepasses/Index.php','/gatepasses.php'=>'Gatepasses/Index.php','/gatepasses/view'=>'Gatepasses/Show.php','/gatepass.php'=>'Gatepasses/Show.php','/gatepasses/create'=>'Gatepasses/Create.php','/gatepass-create.php'=>'Gatepasses/Create.php','/gatepasses/operation'=>'Gatepasses/Operation.php','/gatepass-operation.php'=>'Gatepasses/Operation.php','/gatepasses/cancel'=>'Gatepasses/Cancel.php','/gatepass-cancel.php'=>'Gatepasses/Cancel.php','/gatepasses/operations'=>'Gatepasses/Operations.php','/gate-operations.php'=>'Gatepasses/Operations.php','/gatepasses/qr'=>'Gatepasses/Qr.php','/gatepass-qr.php'=>'Gatepasses/Qr.php','/gatepasses/qr-image'=>'Gatepasses/QrImage.php','/gatepass-qr-image.php'=>'Gatepasses/QrImage.php','/gatepasses/settings'=>'Gatepasses/Settings.php','/gatepass-settings.php'=>'Gatepasses/Settings.php','/gatepasses/types'=>'Gatepasses/Types.php','/gatepass-types.php'=>'Gatepasses/Types.php','/gatepasses/types/edit'=>'Gatepasses/TypeEdit.php','/gatepass-types-edit.php'=>'Gatepasses/TypeEdit.php','/gates'=>'Gates/Index.php','/gates.php'=>'Gates/Index.php','/gates/edit'=>'Gates/Edit.php','/gates-edit.php'=>'Gates/Edit.php',
        '/employees'=>'Employees/Index.php','/employees.php'=>'Employees/Index.php','/employees/view'=>'Employees/Show.php','/employee.php'=>'Employees/Show.php','/employees/create'=>'Employees/Create.php','/employee-create.php'=>'Employees/Create.php',
        '/departments'=>'Departments/Index.php','/departments.php'=>'Departments/Index.php','/departments/edit'=>'Departments/Edit.php','/departments-edit.php'=>'Departments/Edit.php',
        '/approvals'=>'Approvals/Index.php','/approvals.php'=>'Approvals/Index.php','/approvals/view'=>'Approvals/Show.php','/approval.php'=>'Approvals/Show.php','/approvals/decision'=>'Approvals/Decision.php','/approval-decision.php'=>'Approvals/Decision.php','/approvals/workflows'=>'Approvals/Workflows.php','/approval-workflows.php'=>'Approvals/Workflows.php','/approvals/workflows/edit'=>'Approvals/WorkflowEdit.php','/approval-workflow-edit.php'=>'Approvals/WorkflowEdit.php',
        '/admin/users'=>'Admin/Users.php','/users.php'=>'Admin/Users.php','/admin/users/view'=>'Admin/User.php','/user.php'=>'Admin/User.php','/admin/users/create'=>'Admin/CreateUser.php','/create-user.php'=>'Admin/CreateUser.php','/admin/users/edit'=>'Admin/UserEdit.php','/user-edit.php'=>'Admin/UserEdit.php','/admin/roles'=>'Admin/Roles.php','/roles.php'=>'Admin/Roles.php','/admin/roles/edit'=>'Admin/RoleEdit.php','/roles-edit.php'=>'Admin/RoleEdit.php','/admin/roles/permissions'=>'Admin/RolePermissions.php','/role-permissions.php'=>'Admin/RolePermissions.php','/admin/id-types'=>'Admin/IdTypes.php','/id-types.php'=>'Admin/IdTypes.php','/admin/id-types/edit'=>'Admin/IdTypeEdit.php','/id-types-edit.php'=>'Admin/IdTypeEdit.php','/admin/invitations'=>'Admin/Invitations.php','/invitations.php'=>'Admin/Invitations.php','/admin/invitations/create'=>'Admin/InviteUser.php','/invite-user.php'=>'Admin/InviteUser.php',
        '/profile.php'=>'Profile/Index.php','/profile'=>'Profile/Index.php','/change-password'=>'Profile/ChangePassword.php','/change-password.php'=>'Profile/ChangePassword.php','/settings'=>'Settings/Index.php','/settings.php'=>'Settings/Index.php','/settings/theme'=>'Settings/Theme.php','/theme-settings.php'=>'Settings/Theme.php','/theme-settings'=>'Settings/Theme.php','/media'=>'Settings/Media.php','/media.php'=>'Settings/Media.php',
        '/platform/login'=>'Platform/Login.php','/platform-login.php'=>'Platform/Login.php','/platform/tenants'=>'Platform/Tenants.php','/platform.php'=>'Platform/Tenants.php','/platform/tenants/create'=>'Platform/CreateTenant.php','/platform/tenants/view'=>'Platform/TenantView.php','/platform/tenant-view'=>'Platform/TenantView.php','/platform/tenants/edit'=>'Platform/TenantEdit.php','/platform/tenant-edit'=>'Platform/TenantEdit.php','/platform/tenants/database'=>'Platform/TenantDatabase.php','/platform/tenant-database'=>'Platform/TenantDatabase.php',
    ];

    public function dispatch(string $path): void
    {
        $path = rtrim($path, '/') ?: '/';

        // Local tenant development uses /{tenant-slug}/{route}. Strip the
        // tenant prefix only when the remaining path is a real application
        // route, so normal platform routes are not affected.
        if (!isset($this->routes[$path])) {
            $parts = array_values(array_filter(explode('/', trim($path, '/')), static fn ($part) => $part !== ''));
            if (count($parts) >= 2) {
                $tenantSlug = strtolower((string) array_shift($parts));
                $tenantPath = '/' . implode('/', $parts);
                $reserved = ['platform', 'assets', 'api'];
                if (
                    $tenantSlug !== '' &&
                    !in_array($tenantSlug, $reserved, true) &&
                    isset($this->routes[$tenantPath])
                ) {
                    $path = $tenantPath;
                }
            }
        }
        $canonical = [
            '/login.php'=>'/login','/logout.php'=>'/logout','/dashboard.php'=>'/dashboard',
            '/visitors.php'=>'/visitors','/visitor.php'=>'/visitors/view','/visitor-create.php'=>'/visitors/create','/visitor-edit.php'=>'/visitors/edit','/visitor-blacklist.php'=>'/visitors/blacklist','/visitor-companies.php'=>'/visitors/companies','/visitor-companies-edit.php'=>'/visitors/companies/edit','/visitor-settings.php'=>'/visitors/settings',
            '/visits.php'=>'/visits','/visit.php'=>'/visits/view','/visit-create.php'=>'/visits/create','/visit-operation.php'=>'/visits/operation','/visit-types.php'=>'/visits/types','/visit-types-edit.php'=>'/visits/types/edit',
            '/gatepasses.php'=>'/gatepasses','/gatepass.php'=>'/gatepasses/view','/gatepass-create.php'=>'/gatepasses/create','/gatepass-operation.php'=>'/gatepasses/operation','/gatepass-cancel.php'=>'/gatepasses/cancel','/gate-operations.php'=>'/gatepasses/operations','/gatepass-qr.php'=>'/gatepasses/qr','/gatepass-qr-image.php'=>'/gatepasses/qr-image','/gatepass-settings.php'=>'/gatepasses/settings','/gatepass-types.php'=>'/gatepasses/types','/gatepass-types-edit.php'=>'/gatepasses/types/edit','/gates.php'=>'/gates','/gates-edit.php'=>'/gates/edit',
            '/employees.php'=>'/employees','/employee.php'=>'/employees/view','/employee-create.php'=>'/employees/create',
            '/departments.php'=>'/departments','/departments-edit.php'=>'/departments/edit',
            '/approvals.php'=>'/approvals','/approval.php'=>'/approvals/view','/approval-decision.php'=>'/approvals/decision','/approval-workflows.php'=>'/approvals/workflows','/approval-workflow-edit.php'=>'/approvals/workflows/edit',
            '/roles.php'=>'/admin/roles','/roles-edit.php'=>'/admin/roles/edit','/role-permissions.php'=>'/admin/roles/permissions',
            '/users.php'=>'/admin/users','/user.php'=>'/admin/users/view','/create-user.php'=>'/admin/users/create','/user-edit.php'=>'/admin/users/edit',
            '/id-types.php'=>'/admin/id-types','/id-types-edit.php'=>'/admin/id-types/edit','/invitations.php'=>'/admin/invitations','/invite-user.php'=>'/admin/invitations/create',
            '/profile.php'=>'/profile','/change-password.php'=>'/change-password','/settings.php'=>'/settings','/theme-settings.php'=>'/settings/theme','/media.php'=>'/media',
            '/platform-login.php'=>'/platform/login','/platform.php'=>'/platform/tenants','/platform/tenant-view'=>'/platform/tenants/view','/platform/tenant-edit'=>'/platform/tenants/edit','/platform/tenant-database'=>'/platform/tenants/database',
        ];
        if (isset($canonical[$path])) { redirect($canonical[$path]); }
        $target = $this->routes[$path] ?? null;
        if ($target === null) { http_response_code(404); require dirname(__DIR__) . '/modules/Errors/404.php'; return; }
        require dirname(__DIR__) . '/modules/' . $target;
    }
}