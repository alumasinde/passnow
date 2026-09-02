<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();

$errors = [];
$data = [];
try {
    $data = Auth::api(App::api(), 'GET', '/api/v1/theme');
} catch (ApiException $e) {
    $errors[] = $e->getMessage();
} catch (Throwable) {
    $errors[] = 'Unable to load theme settings.';
}

if (requestMethod() === 'POST') {
    Csrf::requireValid($_POST['_csrf'] ?? null);
    $payload = [
        'brand_name' => trim((string)($_POST['brand_name'] ?? '')),
        'logo_url' => trim((string)($_POST['logo_url'] ?? '')),
        'favicon_url' => trim((string)($_POST['favicon_url'] ?? '')),
        'primary_color' => (string)($_POST['primary_color'] ?? ''),
        'primary_color_dark' => (string)($_POST['primary_color_dark'] ?? ''),
        'accent_color' => (string)($_POST['accent_color'] ?? ''),
        'sidebar_background' => (string)($_POST['sidebar_background'] ?? ''),
        'sidebar_text' => (string)($_POST['sidebar_text'] ?? ''),
        'sidebar_active_background' => (string)($_POST['sidebar_active_background'] ?? ''),
        'sidebar_active_text' => (string)($_POST['sidebar_active_text'] ?? ''),
        'appearance' => (string)($_POST['appearance'] ?? 'light'),
    ];

    try {
        $data = Auth::api(App::api(), 'PUT', '/api/v1/theme', $payload);
        Theme::forget();
        flash('success', 'Tenant theme updated successfully.');
        redirect('theme-settings.php');
    } catch (ApiException $e) {
        $errors[] = $e->getMessage();
        $data = $payload;
    } catch (Throwable) {
        $errors[] = 'Unable to save theme settings.';
        $data = $payload;
    }
}

App::render('admin/theme-settings', compact('data', 'errors'));
