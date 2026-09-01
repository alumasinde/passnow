<?php
declare(strict_types=1);

// Router script for PHP's built-in development server.
// Start it from the project root with:
// php -S 0.0.0.0:8000 -t frontend/public frontend/public/router.php

$uri = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
$path = __DIR__ . str_replace('/', DIRECTORY_SEPARATOR, $uri);

if ($uri !== '/' && is_file($path)) {
    return false;
}

require __DIR__ . '/index.php';
