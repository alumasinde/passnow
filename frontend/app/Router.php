<?php
declare(strict_types=1);

final class Router
{
    /** @var array<string,string> */
    private array $routes = [
        '/' => 'login.php',
        '/login' => 'login.php',
        '/logout' => 'logout.php',
        '/dashboard' => 'dashboard.php',
        '/visitors' => 'visitors.php',
        '/visitors/create' => 'visitor-create.php',
        '/visitors/show' => 'visitor.php',
        '/visits' => 'visits.php',
        '/visits/create' => 'visit-create.php',
        '/visits/show' => 'visit.php',
        '/visits/types' => 'visit-types.php',
        '/visits/types/edit' => 'visit-types-edit.php',
        '/gatepasses' => 'gatepasses.php',
        '/gatepasses/create' => 'gatepass-create.php',
        '/gatepasses/show' => 'gatepass.php',
        '/gatepasses/operations' => 'gate-operations.php',
        '/gatepasses/operation' => 'gatepass-operation.php',
        '/gatepasses/qr' => 'gatepass-qr.php',
        '/gatepasses/settings' => 'gatepass-settings.php',
        '/gatepasses/types' => 'gatepass-types.php',
        '/gatepasses/types/edit' => 'gatepass-types-edit.php',
        '/employees' => 'employees.php',
        '/employees/create' => 'employee-create.php',
        '/employees/show' => 'employee.php',
        '/departments' => 'departments.php',
        '/departments/edit' => 'departments-edit.php',
        '/approvals' => 'approvals.php',
        '/approvals/show' => 'approval.php',
        '/approvals/decision' => 'approval-decision.php',
        '/approvals/workflows' => 'approval-workflows.php',
        '/approvals/workflows/edit' => 'approval-workflow-edit.php',
        '/admin/users' => 'users.php',
        '/admin/users/show' => 'user.php',
        '/admin/roles' => 'roles.php',
        '/admin/roles/permissions' => 'role-permissions.php',
        '/admin/id-types' => 'id-types.php',
        '/admin/id-types/edit' => 'id-types-edit.php',
        '/settings' => 'settings.php',
        '/platform/login' => 'platform-login.php',
        '/platform/tenants' => 'platform.php',
    ];

    public function dispatch(string $path): void
    {
        $path = rtrim($path, '/') ?: '/';
        $target = $this->routes[$path] ?? null;

        if ($target === null) {
            http_response_code(404);
            require dirname(__DIR__) . '/public/404.php';
            return;
        }

        require dirname(__DIR__) . '/public/' . $target;
    }
}
