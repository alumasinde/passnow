<?php
declare(strict_types=1);

require_once __DIR__ . '/../../app/App.php';

Auth::requireLogin();

$id = filter_input(INPUT_GET, 'id', FILTER_VALIDATE_INT);
if (!$id) { http_response_code(400); exit('Invalid gatepass ID.'); }

$token = null;
$error = null;

try {
    $payload = Auth::api(App::api(), 'GET', '/api/v1/gatepasses/' . $id);
    $gatepass = apiValue($payload, 'gatepass', $payload['data'] ?? $payload);
    $token = is_array($gatepass) ? ($gatepass['qr_token'] ?? null) : null;
} catch (Throwable $e) {
    $error = $e instanceof ApiException ? $e->getMessage() : 'Unable to load QR details.';
}

App::render('gatepasses/qr', compact('id', 'token', 'error'), 'app');
