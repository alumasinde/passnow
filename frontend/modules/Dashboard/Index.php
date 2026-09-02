<?php
declare(strict_types=1);

require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();

$dashboard = [];
$error = null;

try {
    $dashboard = Auth::api(App::api(), 'GET', '/api/v1/dashboard');
} catch (ApiException $e) {
    $error = $e->getMessage();
} catch (Throwable) {
    $error = 'Dashboard data is temporarily unavailable.';
}

App::render('dashboard/index', [
    'title' => 'Dashboard',
    'dashboard' => $dashboard,
    'error' => $error,
]);
