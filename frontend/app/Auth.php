<?php
declare(strict_types=1);

final class Auth
{
    public static function login(ApiClient $api, string $email, string $password): void
    {
        $r = $api->request('POST', '/api/v1/auth/login', ['email'=>$email,'password'=>$password]);
        if (empty($r['access_token']) || empty($r['refresh_token'])) {
            throw new ApiException('The API did not return a valid session.', 502);
        }
        session_regenerate_id(true);
        $_SESSION['access_token'] = $r['access_token'];
        $_SESSION['refresh_token'] = $r['refresh_token'];
        $_SESSION['user'] = $r['user'] ?? [];
        $_SESSION['tenant_slug'] = strtolower(trim((string)($r['tenant_slug'] ?? localTenantSlug())));
        $_SESSION['authenticated_at'] = time();
        $_SESSION['must_change_password'] = !empty($r['must_change_password']);
        $_SESSION['must_change_password'] = !empty($r['must_change_password']);
    }

    public static function platformLogin(ApiClient $api, string $email, string $password): void
    {
        $r = $api->request('POST', '/api/v1/platform/auth/login', ['email'=>$email,'password'=>$password]);
        if (empty($r['access_token'])) throw new ApiException('The API did not return a valid platform session.', 502);
        session_regenerate_id(true);
        $_SESSION['platform_access_token'] = $r['access_token'];
        $_SESSION['platform_user'] = $r['user'] ?? [];
        $_SESSION['platform_role'] = $r['role'] ?? '';
    }

    public static function check(): bool { return !empty($_SESSION['access_token']); }
    public static function platformCheck(): bool { return !empty($_SESSION['platform_access_token']); }
    public static function requireLogin(): void { if (!self::check()) redirect('login'); if (!empty($_SESSION['must_change_password'])) { $path=parse_url($_SERVER['REQUEST_URI']??'',PHP_URL_PATH)?:''; if (!str_contains($path,'change-password')) redirect('change-password'); } }
    public static function requirePlatform(): void { if (!self::platformCheck()) redirect('platform/login'); }
    public static function user(): array { return is_array($_SESSION['user'] ?? null) ? $_SESSION['user'] : []; }
    public static function platformToken(): ?string { return $_SESSION['platform_access_token'] ?? null; }
    public static function accessToken(): ?string { return $_SESSION['access_token'] ?? null; }

    public static function logout(ApiClient $api): void
    {
        try {
            if (!empty($_SESSION['access_token'])) $api->request('POST', '/api/v1/auth/logout', null, $_SESSION['access_token']);
        } catch (Throwable) {
        }
        self::clearTenantSession();
        session_destroy();
    }

    public static function api(ApiClient $api, string $method, string $path, ?array $body = null): array
    {
        try {
            return $api->request($method, $path, $body, self::accessToken());
        } catch (ApiException $e) {
            if ($e->status() === 401) self::handleExpiredSession();
            throw $e;
        }
    }


    public static function apiBinary(ApiClient $api, string $method, string $path): array
    {
        try { return $api->requestBinary($method, $path, self::accessToken()); }
        catch (ApiException $e) { if ($e->status() === 401) self::handleExpiredSession(); throw $e; }
    }

    public static function apiMultipart(ApiClient $api, string $method, string $path, array $fields, array $files): array
    {
        try {
            return $api->requestMultipart($method, $path, $fields, $files, self::accessToken());
        } catch (ApiException $e) {
            if ($e->status() === 401) self::handleExpiredSession();
            throw $e;
        }
    }

    public static function platformApi(ApiClient $api, string $method, string $path, ?array $body = null): array
    {
        return $api->request($method, $path, $body, self::platformToken());
    }

    private static function handleExpiredSession(): void
    {
        self::clearTenantSession();
        flash('danger', 'Your session has expired. Please sign in again.');
        redirect('login');
    }

    private static function clearTenantSession(): void
    {
        unset($_SESSION['access_token'], $_SESSION['refresh_token'], $_SESSION['user'], $_SESSION['tenant_slug'], $_SESSION['authenticated_at'], $_SESSION['must_change_password']);
    }
}
