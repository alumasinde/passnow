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
        $_SESSION['authenticated_at'] = time();
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
    public static function requireLogin(): void { if (!self::check()) redirect('login'); }
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
        unset($_SESSION['access_token'], $_SESSION['refresh_token'], $_SESSION['user'], $_SESSION['authenticated_at']);
    }
}
