<?php
declare(strict_types=1);

require_once __DIR__ . '/../app/App.php';

Auth::requirePlatform();

$error = '';
$form = [
    'tenant_name' => '',
    'tenant_slug' => '',
    'admin_first_name' => '',
    'admin_last_name' => '',
    'admin_email' => '',
    'database_mode' => 'existing',
    'database_host' => '127.0.0.1',
    'database_port' => '3306',
    'database_name' => '',
    'database_username' => '',
];

if (requestMethod() === 'POST') {
    foreach ($form as $key => $_) {
        $form[$key] = trim((string) ($_POST[$key] ?? ''));
    }
    $password = (string) ($_POST['admin_password'] ?? '');
    $confirmPassword = (string) ($_POST['admin_password_confirmation'] ?? '');

    try {
        Csrf::requireValid($_POST['_csrf'] ?? null);

        if ($password !== $confirmPassword) {
            throw new InvalidArgumentException('Passwords do not match.');
        }

        Auth::platformApi(App::api(), 'POST', '/api/v1/platform/tenants', [
            'tenant_name' => $form['tenant_name'],
            'tenant_slug' => strtolower($form['tenant_slug']),
            'admin_first_name' => $form['admin_first_name'],
            'admin_last_name' => $form['admin_last_name'],
            'admin_email' => $form['admin_email'],
            'admin_password' => $password,
            'database_mode' => $form['database_mode'],
            'database_host' => $form['database_host'],
            'database_port' => $form['database_port'],
            'database_name' => $form['database_name'],
            'database_username' => $form['database_username'],
            'database_password' => (string) ($_POST['database_password'] ?? ''),
        ]);

        flash('success', 'Organization and its first administrator were created successfully.');
        redirect('platform/tenants');
    } catch (Throwable $e) {
        $error = $e->getMessage();
    }
}

App::render('admin/tenants/create', [
    'form' => $form,
    'error' => $error,
], 'platform');
