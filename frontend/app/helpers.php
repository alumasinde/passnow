<?php
declare(strict_types=1);
final class AppContext { private static array $config=[]; public static function init(array $config):void{self::$config=$config;} public static function config(?string $key=null,mixed $default=null):mixed{if($key===null)return self::$config;$v=self::$config;foreach(explode('.',$key) as $part){if(!is_array($v)||!array_key_exists($part,$v))return $default;$v=$v[$part];}return $v;} }
function e(mixed $value):string{
    if (is_string($value) && preg_match('/^\\d{4}-\\d{2}-\\d{2}[ T]\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?(?:Z|[+-]\\d{2}:?\\d{2})?$/', $value)) {
        $value = FormatDate($value);
    }
    return htmlspecialchars((string)$value, ENT_QUOTES|ENT_SUBSTITUTE, 'UTF-8');
}
function localTenantSlug(): string
{
    $path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
    $parts = array_values(array_filter(explode('/', trim($path, '/')), static fn ($part) => $part !== ''));
    if ($parts === []) return '';

    $candidate = strtolower((string) $parts[0]);
    $reserved = ['login', 'logout', 'dashboard', 'platform', 'assets', 'api', 'admin', 'settings'];
    if (in_array($candidate, $reserved, true) || str_contains($candidate, '.php')) return '';

    return $candidate;
}

function url(string $path=''):string
{
    $base = rtrim((string) AppContext::config('app.base_url', ''), '/');
    $path = '/' . ltrim($path, '/');

    // Keep the tenant prefix on internal navigation during local development.
    $tenantSlug = localTenantSlug();
    if ($tenantSlug !== '' && !str_starts_with($path, '/platform')) {
        $path = '/' . $tenantSlug . $path;
    }

    return $base . $path;
}

function asset(string $path):string
{
    $base = rtrim((string) AppContext::config('app.base_url', ''), '/');
    return $base . '/assets/' . ltrim($path, '/');
}
function redirect(string $path):never{header('Location: '.(preg_match('~^https?://~i',$path)?$path:url($path)));exit;}
function requestMethod():string{return strtoupper($_SERVER['REQUEST_METHOD']??'GET');}
function flash(string $type,string $message):void{$_SESSION['_flash'][]=['type'=>$type,'message'=>$message];}
function consumeFlash():array{$v=$_SESSION['_flash']??[];unset($_SESSION['_flash']);return is_array($v)?$v:[];}


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

function FormatDate(mixed $value, string $format = 'Y-m-d: H:i'): string
{
    if ($value === null || $value === '') return '—';
    try {
        if ($value instanceof DateTimeInterface) return $value->format($format);
        return (new DateTimeImmutable((string)$value))->format($format);
    } catch (Throwable) {
        return (string)$value;
    }
}
