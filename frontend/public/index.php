<?php
declare(strict_types=1);

require_once __DIR__ . '/../../app/App.php';
require_once __DIR__ . '/../app/Router.php';

$path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
$base = rtrim(dirname($_SERVER['SCRIPT_NAME'] ?? ''), '/');
if ($base !== '' && $base !== '/' && str_starts_with($path, $base)) {
    $path = substr($path, strlen($base)) ?: '/';
}

(new Router())->dispatch($path);
