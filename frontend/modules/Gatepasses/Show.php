<?php
declare(strict_types=1);

require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();

$id = filter_input(INPUT_GET, 'id', FILTER_VALIDATE_INT);
if (!$id) {
    http_response_code(400);
    exit('Invalid gatepass ID.');
}

$gatepass = [];
$movements = [];
$error = null;

try {
    $payload = Auth::api(App::api(), 'GET', '/api/v1/gatepasses/' . rawurlencode((string)$id));
    $gatepass = apiValue($payload, 'gatepass', $payload['data'] ?? $payload);
    if (!is_array($gatepass)) $gatepass = [];
} catch (ApiException $e) {
    $error = $e->getMessage();
}

try {
    $movementPayload = Auth::api(App::api(), 'GET', '/api/v1/gatepasses/' . rawurlencode((string)$id) . '/movements');
    $movements = apiRows($movementPayload);
} catch (Throwable) {
    // Movement history is supplementary to the main record.
}

if (!$gatepass && !$error) {
    http_response_code(404);
    $error = 'Gatepass not found.';
}

App::render('gatepasses/show', [
    'title' => 'Gatepass',
    'gatepass' => $gatepass,
    'movements' => $movements,
    'error' => $error,
]);
