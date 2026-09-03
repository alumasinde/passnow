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

    // Central URL normalization keeps legacy module filenames out of links.
    // Query strings are preserved, so existing id-based detail routes continue
    // to work while the browser always receives a clean path.
    $parts = parse_url($path);
    $pathname = (string)($parts['path'] ?? $path);
    $query = isset($parts['query']) ? '?' . $parts['query'] : '';
    $clean = [
        '/login.php'=>'/login','/logout.php'=>'/logout','/dashboard.php'=>'/dashboard',
        '/visitors.php'=>'/visitors','/visitor.php'=>'/visitors/view','/visitor-create.php'=>'/visitors/create','/visitor-edit.php'=>'/visitors/edit','/visitor-blacklist.php'=>'/visitors/blacklist','/visitor-companies.php'=>'/visitors/companies','/visitor-companies-edit.php'=>'/visitors/companies/edit','/visitor-settings.php'=>'/visitors/settings',
        '/visits.php'=>'/visits','/visit.php'=>'/visits/view','/visit-create.php'=>'/visits/create','/visit-operation.php'=>'/visits/operation','/visit-types.php'=>'/visits/types','/visit-types-edit.php'=>'/visits/types/edit',
        '/gatepasses.php'=>'/gatepasses','/gatepass.php'=>'/gatepasses/view','/gatepass-create.php'=>'/gatepasses/create','/gatepass-operation.php'=>'/gatepasses/operation','/gatepass-cancel.php'=>'/gatepasses/cancel','/gate-operations.php'=>'/gatepasses/operations','/gatepass-qr.php'=>'/gatepasses/qr','/gatepass-qr-image.php'=>'/gatepasses/qr-image','/gatepass-settings.php'=>'/gatepasses/settings','/gatepass-types.php'=>'/gatepasses/types','/gatepass-types-edit.php'=>'/gatepasses/types/edit','/gates.php'=>'/gates','/gates-edit.php'=>'/gates/edit',
        '/employees.php'=>'/employees','/employee.php'=>'/employees/view','/employee-create.php'=>'/employees/create',
        '/departments.php'=>'/departments','/departments-edit.php'=>'/departments/edit',
        '/approvals.php'=>'/approvals','/approval.php'=>'/approvals/view','/approval-decision.php'=>'/approvals/decision','/approval-workflows.php'=>'/approvals/workflows','/approval-workflow-edit.php'=>'/approvals/workflows/edit',
        '/users.php'=>'/admin/users','/user.php'=>'/admin/users/view','/create-user.php'=>'/admin/users/create','/user-edit.php'=>'/admin/users/edit',
        '/roles.php'=>'/admin/roles','/roles-edit.php'=>'/admin/roles/edit','/role-permissions.php'=>'/admin/roles/permissions','/id-types.php'=>'/admin/id-types','/id-types-edit.php'=>'/admin/id-types/edit','/invitations.php'=>'/admin/invitations','/invite-user.php'=>'/admin/invitations/create',
        '/profile.php'=>'/profile','/change-password.php'=>'/change-password','/settings.php'=>'/settings','/theme-settings.php'=>'/settings/theme','/media.php'=>'/media',
        '/platform-login.php'=>'/platform/login','/platform.php'=>'/platform/tenants',
    ];
    $path = ($clean[$pathname] ?? $pathname) . $query;

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
    // Several Go handlers intentionally return a plain JSON array. Do not
    // mistake that successful response for an empty result set.
    if (array_is_list($payload)) return $payload;
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
