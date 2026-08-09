<?php
declare(strict_types=1);

return [
    'app' => [
        'name' => getenv('APP_NAME') ?: 'PassNow',
        'environment' => getenv('APP_ENV') ?: 'development',
        'base_url' => rtrim(getenv('APP_BASE_URL') ?: '', '/'),
        'api_base_url' => rtrim(getenv('API_BASE_URL') ?: 'http://127.0.0.1:8080', '/'),
        'tenant_host' => getenv('TENANT_HOST') ?: ($_SERVER['HTTP_HOST'] ?? ''),
        'timezone' => getenv('APP_TIMEZONE') ?: 'Africa/Nairobi',
        'session_name' => getenv('SESSION_NAME') ?: 'passnow_session',
    ],
    'security' => [
        'session_secure' => filter_var(getenv('SESSION_SECURE') ?: '0', FILTER_VALIDATE_BOOL),
        'session_samesite' => getenv('SESSION_SAMESITE') ?: 'Lax',
        'csrf_ttl' => (int) (getenv('CSRF_TTL') ?: 3600),
        'api_timeout' => (int) (getenv('API_TIMEOUT') ?: 15),
    ],
    'ui' => [
        'font_awesome_url' => getenv('FONT_AWESOME_URL') ?: 'https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.7.2/css/all.min.css',
        'page_size' => (int) (getenv('PAGE_SIZE') ?: 20),
    ],
];
