<?php
declare(strict_types=1);

require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();

$query = new ListQuery(['status', 'type']);
$page = $query->page();
$perPage = $query->perPage((int) App::config('ui.page_size', 20));

$params = [
    'page' => $page,
    'per_page' => $perPage,
];
if ($query->search() !== '') $params['q'] = $query->search();
if (!empty($_GET['status'])) $params['status'] = trim((string) $_GET['status']);
if (!empty($_GET['type'])) $params['type'] = trim((string) $_GET['type']);

$rows = [];
$total = 0;
$error = null;
$meta = [];

try {
    $result = Auth::api(
        App::api(),
        'GET',
        '/api/v1/gatepasses?' . http_build_query($params)
    );

    $rows = $result['data'] ?? $result['items'] ?? $result['results'] ?? [];
    $rows = is_array($rows) ? $rows : [];

    $meta = is_array($result['meta'] ?? null) ? $result['meta'] : [];
    $total = (int) ($meta['total'] ?? $result['total'] ?? count($rows));
} catch (ApiException $e) {
    $error = $e->getMessage();
} catch (Throwable) {
    $error = 'Gatepass data is temporarily unavailable.';
}

$paginator = new Paginator($page, $perPage, $total);

$statusOptions = is_array($meta['statuses'] ?? null) ? $meta['statuses'] : [];
$typeOptions = is_array($meta['types'] ?? null) ? $meta['types'] : [];

$columns = [
    ['key' => 'gatepass_number', 'label' => 'Number'],
    ['key' => 'gatepass_type_name', 'label' => 'Type'],
    ['key' => 'subject_name', 'label' => 'Person'],
    ['key' => 'status', 'label' => 'Status'],
    ['key' => 'created_at', 'label' => 'Created'],
];

$rowActions = [
    [
        'label' => 'View',
        'icon' => 'fa-eye',
        'href' => static fn(array $row): string =>
            url('gatepass.php?id=' . rawurlencode((string) ($row['id'] ?? ''))),
    ],
];

App::render('gatepasses/index', compact(
    'query', 'rows', 'paginator', 'columns', 'rowActions',
    'error', 'statusOptions', 'typeOptions'
));
