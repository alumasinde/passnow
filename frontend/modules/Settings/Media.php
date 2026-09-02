<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();

$errors = [];
$items = [];

if (requestMethod() === 'POST') {
    Csrf::requireValid($_POST['_csrf'] ?? null);

    try {
        $action = (string)($_POST['action'] ?? '');
        if ($action === 'upload') {
            if (!isset($_FILES['file']) || (int)($_FILES['file']['error'] ?? UPLOAD_ERR_NO_FILE) !== UPLOAD_ERR_OK) {
                throw new RuntimeException('Please select a file to upload.');
            }
            Auth::apiMultipart(
                App::api(),
                'POST',
                '/api/v1/media',
                ['purpose' => trim((string)($_POST['purpose'] ?? 'general')) ?: 'general'],
                ['file' => (string)$_FILES['file']['tmp_name']]
            );
            flash('success', 'Media file uploaded successfully.');
            redirect('media');
        }

        if ($action === 'delete') {
            $id = (int)($_POST['id'] ?? 0);
            if ($id <= 0) throw new RuntimeException('Invalid media file.');
            Auth::api(App::api(), 'DELETE', '/api/v1/media/' . $id);
            flash('success', 'Media file deleted.');
            redirect('media');
        }
    } catch (ApiException $e) {
        $errors[] = $e->getMessage();
    } catch (Throwable $e) {
        $errors[] = $e->getMessage() !== '' ? $e->getMessage() : 'Unable to complete the media operation.';
    }
}

try {
    $response = Auth::api(App::api(), 'GET', '/api/v1/media');
    $items = is_array($response['items'] ?? null) ? $response['items'] : [];
} catch (ApiException $e) {
    $errors[] = $e->getMessage();
} catch (Throwable) {
    $errors[] = 'Unable to load the media library.';
}

App::render('admin/media', compact('items', 'errors'));
