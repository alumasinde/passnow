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
        '/visitors' => 'Visitors/Index.php','/visitors.php'=>'Visitors/Index.php','/visitor.php'=>'Visitors/Show.php','/visitor-create.php'=>'Visitors/Create.php','/visitor-edit.php'=>'Visitors/Edit.php','/visitor-blacklist.php'=>'Visitors/Blacklist.php','/visitor-companies.php'=>'Visitors/Companies.php','/visitor-companies-edit.php'=>'Visitors/CompanyEdit.php','/visitor-settings.php'=>'Visitors/Settings.php',
        '/visits'=>'Visits/Index.php','/visits.php'=>'Visits/Index.php','/visit.php'=>'Visits/Show.php','/visit-create.php'=>'Visits/Create.php','/visit-operation.php'=>'Visits/Operation.php','/visit-types.php'=>'Visits/Types.php','/visit-types-edit.php'=>'Visits/TypeEdit.php',
        '/gatepasses'=>'Gatepasses/Index.php','/gatepasses.php'=>'Gatepasses/Index.php','/gatepass.php'=>'Gatepasses/Show.php','/gatepass-create.php'=>'Gatepasses/Create.php','/gatepass-operation.php'=>'Gatepasses/Operation.php','/gatepass-cancel.php'=>'Gatepasses/Cancel.php','/gate-operations.php'=>'Gatepasses/Operations.php','/gatepass-qr.php'=>'Gatepasses/Qr.php','/gatepass-qr-image.php'=>'Gatepasses/QrImage.php','/gatepass-settings.php'=>'Gatepasses/Settings.php','/gatepass-types.php'=>'Gatepasses/Types.php','/gatepass-types-edit.php'=>'Gatepasses/TypeEdit.php',
        '/employees'=>'Employees/Index.php','/employees.php'=>'Employees/Index.php','/employee.php'=>'Employees/Show.php','/employee-create.php'=>'Employees/Create.php',
        '/departments'=>'Departments/Index.php','/departments.php'=>'Departments/Index.php','/departments-edit.php'=>'Departments/Edit.php',
        '/approvals'=>'Approvals/Index.php','/approvals.php'=>'Approvals/Index.php','/approval.php'=>'Approvals/Show.php','/approval-decision.php'=>'Approvals/Decision.php','/approval-workflows.php'=>'Approvals/Workflows.php','/approval-workflow-edit.php'=>'Approvals/WorkflowEdit.php',
        '/admin/users'=>'Admin/Users.php','/users.php'=>'Admin/Users.php','/user.php'=>'Admin/User.php','/admin/roles'=>'Admin/Roles.php','/roles.php'=>'Admin/Roles.php','/roles-edit.php'=>'Admin/RoleEdit.php','/role-permissions.php'=>'Admin/RolePermissions.php','/id-types.php'=>'Admin/IdTypes.php','/id-types-edit.php'=>'Admin/IdTypeEdit.php','/invitations.php'=>'Admin/Invitations.php','/invite-user.php'=>'Admin/InviteUser.php','/create-user.php'=>'Admin/CreateUser.php',
        '/profile.php'=>'Profile/Index.php','/profile'=>'Profile/Index.php','/change-password'=>'Profile/ChangePassword.php','/change-password.php'=>'Profile/ChangePassword.php','/settings'=>'Settings/Index.php','/settings.php'=>'Settings/Index.php','/theme-settings.php'=>'Settings/Theme.php','/theme-settings'=>'Settings/Theme.php','/media'=>'Settings/Media.php','/media.php'=>'Settings/Media.php',
        '/platform/login'=>'Platform/Login.php','/platform-login.php'=>'Platform/Login.php','/platform/tenants'=>'Platform/Tenants.php','/platform.php'=>'Platform/Tenants.php','/platform/tenants/create'=>'Platform/CreateTenant.php','/platform/tenant-view'=>'Platform/TenantView.php','/platform/tenant-edit'=>'Platform/TenantEdit.php','/platform/tenant-database'=>'Platform/TenantDatabase.php',
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
            '/login.php' => '/login', '/logout.php' => '/logout', '/dashboard.php' => '/dashboard',
            '/visitors.php' => '/visitors', '/visits.php' => '/visits', '/gatepasses.php' => '/gatepasses',
            '/employees.php' => '/employees', '/departments.php' => '/departments', '/approvals.php' => '/approvals',
            '/roles.php' => '/admin/roles', '/users.php' => '/admin/users', '/settings.php' => '/settings',
            '/platform-login.php' => '/platform/login', '/platform.php' => '/platform/tenants',
        ];
        if (isset($canonical[$path])) { redirect($canonical[$path]); }
        $target = $this->routes[$path] ?? null;
        if ($target === null) { http_response_code(404); require dirname(__DIR__) . '/modules/Errors/404.php'; return; }
        require dirname(__DIR__) . '/modules/' . $target;
    }
}