<?php
declare(strict_types=1);

$environment = strtolower(trim((string) (getenv('APP_ENV') ?: 'development')));
$sessionSecureEnv = getenv('SESSION_SECURE');
$sessionSecure = $sessionSecureEnv === false
    ? $environment !== 'development'
    : filter_var($sessionSecureEnv, FILTER_VALIDATE_BOOL);

$sessionSameSite = ucfirst(strtolower(trim((string) (getenv('SESSION_SAMESITE') ?: 'Lax')));
if (!in_array($sessionSameSite, ['Lax', 'Strict', 'None'], true)) {
    $sessionSameSite = 'Lax';
}
if ($sessionSameSite === 'None' && !$sessionSecure) {
    $sessionSameSite = 'Lax';
}

$sessionIdleTimeout = max(300, (int) (getenv('SESSION_IDLE_TIMEOUT') ?: 1800));
$sessionAbsoluteTimeout = max($sessionIdleTimeout, (int) (getenv('SESSION_ABSOLUTE_TIMEOUT') ?: 28800));

return [
    'app' => [
        'name' => getenv('APP_NAME') ?: 'PassNow',
        'environment' => $environment,
        'base_url' => rtrim(getenv('APP_BASE_URL') ?: '', '/'),
        'api_base_url' => rtrim(getenv('API_BASE_URL') ?: 'http://127.0.0.1:8080', '/'),
        'tenant_host' => trim((string) (getenv('TENANT_HOST') ?: ($_SERVER['HTTP_HOST'] ?? ''))),
        'timezone' => getenv('APP_TIMEZONE') ?: 'Africa/Nairobi',
        'session_name' => getenv('SESSION_NAME') ?: 'passnow_session',
    ],
    'security' => [
        'session_secure' => $sessionSecure,
        'session_samesite' => $sessionSameSite,
        'session_idle_timeout' => $sessionIdleTimeout,
        'session_absolute_timeout' => $sessionAbsoluteTimeout,
        'csrf_ttl' => max(300, (int) (getenv('CSRF_TTL') ?: 3600)),
        'api_timeout' => max(3, min(120, (int) (getenv('API_TIMEOUT') ?: 15))),
    ],
    'ui' => [
        'font_awesome_url' => getenv('FONT_AWESOME_URL') ?: 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css',
        'page_size' => max(1, min(100, (int) (getenv('PAGE_SIZE') ?: 20))),
    ],
];
