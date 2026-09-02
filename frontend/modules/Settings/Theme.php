<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';

Auth::requireLogin();

$isAjax = strtolower((string)($_SERVER['HTTP_X_REQUESTED_WITH'] ?? '')) === 'xmlhttprequest';
$json = static function (int $status, array $payload): never {
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode($payload, JSON_UNESCAPED_SLASHES);
    exit;
};

$errors = [];
$data = [];
try {
    $response = Auth::api(App::api(), 'GET', '/api/v1/theme');
    $data = is_array($response['data'] ?? null) ? $response['data'] : $response;
    if (!is_array($data)) $data = [];
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
        foreach (['logo_file' => 'logo_url', 'favicon_file' => 'favicon_url'] as $input => $target) {
            if (isset($_FILES[$input]) && (int)($_FILES[$input]['error'] ?? UPLOAD_ERR_NO_FILE) !== UPLOAD_ERR_NO_FILE) {
                if ((int)($_FILES[$input]['error'] ?? UPLOAD_ERR_OK) !== UPLOAD_ERR_OK) {
                    throw new RuntimeException('Unable to receive the selected media file.');
                }
                $uploaded = Auth::apiMultipart(
                    App::api(),
                    'POST',
                    '/api/v1/media',
                    ['purpose' => 'branding'],
                    ['file' => (string)$_FILES[$input]['tmp_name']]
                );
                if (!empty($uploaded['public_url'])) $payload[$target] = (string)$uploaded['public_url'];
            }
        }

        $response = Auth::api(App::api(), 'PUT', '/api/v1/theme', $payload);
        $data = is_array($response['data'] ?? null) ? $response['data'] : $response;
        if (!is_array($data)) $data = $payload;
        Theme::forget();

        if ($isAjax) {
            $json(200, ['ok' => true, 'message' => 'Tenant theme saved.', 'theme' => $data]);
        }

        flash('success', 'Tenant theme updated successfully.');
        redirect('theme-settings.php');
    } catch (ApiException $e) {
        $errors[] = $e->getMessage();
        $data = array_merge($data, $payload);
    } catch (Throwable $e) {
        $errors[] = $e->getMessage() ?: 'Unable to save theme settings.';
        $data = array_merge($data, $payload);
    }

    if ($isAjax) {
        $json(422, [
            'ok' => false,
            'message' => $errors[0] ?? 'Unable to save theme settings.',
            'errors' => $errors,
        ]);
    }
}

App::render('admin/theme-settings', compact('data', 'errors'));
