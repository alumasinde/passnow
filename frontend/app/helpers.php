<?php
declare(strict_types=1);

final class AppContext {
    private static array $config = [];

    public static function init(array $config): void
    {
        self::$config = $config;
    }

    public static function config(?string $key = null, mixed $default = null): mixed
    {
        if ($key === null) return self::$config;
        $value = self::$config;
        foreach (explode('.', $key) as $part) {
            if (!is_array($value) || !array_key_exists($part, $value)) return $default;
            $value = $value[$part];
        }
        return $value;
    }
}

function e(mixed $value): string
{
    return htmlspecialchars((string) $value, ENT_QUOTES | ENT_SUBSTITUTE, 'UTF-8');
}

function url(string $path = ''): string
{
    $base = (string) AppContext::config('app.base_url', '');
    return $base . '/' . ltrim($path, '/');
}

function asset(string $path): string
{
    return url('assets/' . ltrim($path, '/'));
}

/**
 * Redirect only to an internal application path.
 * External redirects must be explicitly trusted by the caller.
 */
function redirect(string $path): never
{
    if ($path === '' || preg_match('~^(?:[a-z][a-z0-9+.-]*:|//)~i', $path)) {
        $path = 'index.php';
    }

    header('Location: ' . url($path), true, 303);
    exit;
}

function redirectExternal(string $url): never
{
    if (!filter_var($url, FILTER_VALIDATE_URL) || !preg_match('~^https://~i', $url)) {
        redirect('index.php');
    }

    header('Location: ' . $url, true, 303);
    exit;
}

function requestMethod(): string
{
    return strtoupper($_SERVER['REQUEST_METHOD'] ?? 'GET');
}

function flash(string $type, string $message): void
{
    $_SESSION['_flash'][] = ['type' => $type, 'message' => $message];
}

function consumeFlash(): array
{
    $value = $_SESSION['_flash'] ?? [];
    unset($_SESSION['_flash']);
    return is_array($value) ? $value : [];
}

function apiRows(array $payload): array
{
    foreach (['data', 'items', 'results', 'records'] as $key) {
        if (isset($payload[$key]) && is_array($payload[$key])) return $payload[$key];
    }
    return [];
}

function apiMeta(array $payload): array
{
    return is_array($payload['meta'] ?? null) ? $payload['meta'] : [];
}

function apiValue(array $payload, string $key, mixed $default = null): mixed
{
    if (array_key_exists($key, $payload)) return $payload[$key];
    foreach (['data', 'result'] as $container) {
        if (is_array($payload[$container] ?? null) && array_key_exists($key, $payload[$container])) {
            return $payload[$container][$key];
        }
    }
    return $default;
}

function oldOr(string $key, mixed $fallback = ''): mixed
{
    return array_key_exists($key, $_POST) ? $_POST[$key] : $fallback;
}
