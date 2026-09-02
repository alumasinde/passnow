<?php
declare(strict_types=1);

require_once __DIR__ . '/../../app/App.php';

Auth::requireLogin();

$errors = [];
$types = [];
$people = [];
$meta = [];

try {
    $typePayload = Auth::api(App::api(), 'GET', '/api/v1/gatepass-types');
    $types = apiRows($typePayload);
    $meta = apiMeta($typePayload);
} catch (Throwable) {
    // The create form can still render; the API remains authoritative.
}

if (requestMethod() === 'POST') {
    Csrf::requireValid($_POST['_csrf'] ?? null);

    $payload = [
        'gatepass_type_id' => (int)($_POST['gatepass_type_id'] ?? 0),
        'subject_type' => trim((string)($_POST['subject_type'] ?? '')),
        'subject_id' => (int)($_POST['subject_id'] ?? 0),
        'purpose' => trim((string)($_POST['purpose'] ?? '')),
        'direction' => trim((string)($_POST['direction'] ?? '')),
        'returnable' => isset($_POST['returnable']),
        'expected_return_at' => trim((string)($_POST['expected_return_at'] ?? '')),
        'notes' => trim((string)($_POST['notes'] ?? '')),
    ];

    if ($payload['gatepass_type_id'] < 1) $errors[] = 'Select a gatepass type.';
    if (!in_array($payload['subject_type'], ['employee', 'visitor'], true)) $errors[] = 'Select a valid person type.';
    if ($payload['subject_id'] < 1) $errors[] = 'Select a person.';
    if ($payload['purpose'] === '') $errors[] = 'Purpose is required.';
    if (!in_array($payload['direction'], ['out', 'in'], true)) $errors[] = 'Select a valid direction.';
    if ($payload['returnable'] && $payload['expected_return_at'] === '') $errors[] = 'Expected return time is required for returnable gatepasses.';

    if (!$errors) {
        try {
            $created = Auth::api(App::api(), 'POST', '/api/v1/gatepasses', $payload);
            $id = apiValue($created, 'id');

            flash('success', 'Gatepass created successfully.');
            if ($id) redirect('gatepass.php?id=' . rawurlencode((string)$id));
            redirect('gatepasses.php');
        } catch (ApiException $e) {
            $errors[] = $e->getMessage();
        } catch (Throwable) {
            $errors[] = 'Unable to create the gatepass right now.';
        }
    }
}

App::render('gatepasses/create', [
    'title' => 'New Gatepass',
    'errors' => $errors,
    'types' => $types,
    'meta' => $meta,
]);
