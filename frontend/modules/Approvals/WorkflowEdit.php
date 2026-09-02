<?php
declare(strict_types=1);

require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$id = filter_input(INPUT_GET, 'id', FILTER_VALIDATE_INT);
$item = [];
$errors = [];
$roles = [];
$users = [];
$departments = [];

try {
    $rolesPayload = Auth::api(App::api(), 'GET', '/api/v1/roles');
    $roles = apiRows($rolesPayload);
} catch (Throwable) {}

try {
    $departmentsPayload = Auth::api(App::api(), 'GET', '/api/v1/departments');
    $departments = apiRows($departmentsPayload);
} catch (Throwable) {}

try {
    $usersPayload = Auth::api(App::api(), 'GET', '/api/v1/users');
    $users = array_map(static function(array $u): array { $u['id']=(int)($u['user_id']??$u['id']??0); $u['name']=trim((string)($u['first_name']??'').' '.(string)($u['last_name']??'')); if($u['name']==='')$u['name']=(string)($u['email']??''); return $u; }, apiRows($usersPayload));
} catch (Throwable) {}

if ($id) {
    try {
        $payload = Auth::api(App::api(), 'GET', '/api/v1/approval-workflows/' . $id);
        $item = apiValue($payload, 'workflow', $payload['data'] ?? $payload);
        if (!is_array($item)) $item = [];
    } catch (ApiException $e) {
        $errors[] = $e->getMessage();
    } catch (Throwable) {
        $errors[] = 'Unable to load workflow.';
    }
}

if (requestMethod() === 'POST') {
    Csrf::requireValid($_POST['_csrf'] ?? null);

    $steps = [];
    foreach ((array)($_POST['steps'] ?? []) as $step) {
        if (!is_array($step)) continue;
        $type = trim((string)($step['approver_type'] ?? 'role'));
        $row = [
            'label' => trim((string)($step['label'] ?? '')),
            'approver_type' => $type,
            'required' => !empty($step['required']),
        ];
        $approverID = (int)($step['approver_id'] ?? 0);
        if ($type === 'role') {
            $row['role_id'] = $approverID ?: null;
            $row['user_id'] = null;
        } else {
            $row['user_id'] = $approverID ?: null;
            $row['role_id'] = null;
        }
        $steps[] = $row;
    }

    $payload = [
        'name' => trim((string)($_POST['workflow_name'] ?? $_POST['name'] ?? '')),
        'active' => !empty($_POST['active']),
        'steps' => $steps,
    ];

    if ($payload['name'] === '') $errors[] = 'Workflow name is required.';
    if (!$steps) $errors[] = 'Add at least one approval step.';

    foreach ($steps as $index => $step) {
        if ($step['label'] === '' || (($step['role_id'] ?? null) === null && ($step['user_id'] ?? null) === null)) {
            $errors[] = 'Every step needs a name and an approver.';
            break;
        }
    }

    if (!$errors) {
        try {
            if ($id) {
                Auth::api(App::api(), 'PATCH', '/api/v1/approval-workflows/' . $id, $payload);
            } else {
                Auth::api(App::api(), 'POST', '/api/v1/approval-workflows', $payload);
            }
            flash('success', 'Approval workflow saved.');
            redirect('approval-workflows.php');
        } catch (ApiException $e) {
            $errors[] = $e->getMessage();
        } catch (Throwable) {
            $errors[] = 'Unable to save workflow.';
        }
    }

    $item = $payload;
}

$steps = is_array($item['steps'] ?? null) ? $item['steps'] : [];
if (!$steps) {
    $steps = [[
        'label' => '',
        'approver_type' => 'role',
        'role_id' => null,
        'user_id' => null,
        'required' => true,
    ]];
}

App::render('admin/approval-workflow-edit', compact('id', 'item', 'errors', 'roles', 'users', 'departments', 'steps'));
