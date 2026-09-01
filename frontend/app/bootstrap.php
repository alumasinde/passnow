<?php
declare(strict_types=1);

$envPath = dirname(__DIR__, 2) . DIRECTORY_SEPARATOR . '.env';
if (is_file($envPath) && is_readable($envPath)) {
    foreach (file($envPath, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) ?: [] as $line) {
        $line = trim($line);
        if ($line === '' || str_starts_with($line, '#')) continue;
        $parts = explode('=', $line, 2);
        if (count($parts) !== 2) continue;
        [$key, $value] = $parts;
        $key = trim($key);
        $value = trim($value);
        if ($key === '' || getenv($key) !== false) continue;
        if ((str_starts_with($value, '"') && str_ends_with($value, '"')) || (str_starts_with($value, "'") && str_ends_with($value, "'"))) $value = substr($value, 1, -1);
        putenv($key . '=' . $value);
        $_ENV[$key] = $value;
    }
}

$config = require __DIR__ . '/../config/config.php';
date_default_timezone_set($config['app']['timezone']);
ini_set('session.use_strict_mode', '1');
ini_set('session.use_only_cookies', '1');
ini_set('session.cookie_httponly', '1');
ini_set('session.cookie_secure', $config['security']['session_secure'] ? '1' : '0');
ini_set('session.cookie_samesite', $config['security']['session_samesite']);
session_name($config['app']['session_name']);
session_start();

require_once __DIR__ . '/helpers.php';
require_once __DIR__ . '/Csrf.php';
require_once __DIR__ . '/ApiException.php';
require_once __DIR__ . '/ApiClient.php';
require_once __DIR__ . '/Auth.php';
require_once __DIR__ . '/View.php';
require_once __DIR__ . '/ListQuery.php';
require_once __DIR__ . '/Paginator.php';

AppContext::init($config);
