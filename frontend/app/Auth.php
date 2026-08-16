<?php
declare(strict_types=1);

final class Auth
{
    public static function login(ApiClient $api, string $email, string $password): void
    {
        $response = $api->request('POST', '/api/v1/auth/login', [
            'email' => $email,
            'password' => $password,
        ]);

        if (empty($response['access_token']) || empty($response['refresh_token'])) {
            throw new ApiException('The API did not return a valid session.', 502);
        }

        session_regenerate_id(true);
        $_SESSION['access_token'] = (string) $response['access_token'];
        $_SESSION['refresh_token'] = (string) $response['refresh_token'];
        $_SESSION['user'] = is_array($response['user'] ?? null) ? $response['user'] : [];
        $_SESSION['authenticated_at'] = time();
        $_SESSION['last_activity_at'] = time();
        Csrf::rotate();
    }

    public static function logout(ApiClient $api): void
    {
        try {
            if (!empty($_SESSION['access_token'])) {
                $api->request('POST', '/api/v1/auth/logout', null, (string) $_SESSION['access_token']);
            }
        } catch (Throwable) {
            // Local session cleanup must still happen when the API is unavailable.
        }

        self::destroyLocalSession();
    }

    public static function check(): bool
    {
        if (empty($_SESSION['access_token'])) return false;

        $now = time();
        $authenticatedAt = (int) ($_SESSION['authenticated_at'] ?? 0);
        $lastActivityAt = (int) ($_SESSION['last_activity_at'] ?? $authenticatedAt);
        $absoluteTimeout = (int) AppContext::config('security.session_absolute_timeout', 28800);
        $idleTimeout = (int) AppContext::config('security.session_idle_timeout', 1800);

        if ($authenticatedAt <= 0 || ($now - $authenticatedAt) > $absoluteTimeout || ($now - $lastActivityAt) > $idleTimeout) {
            self::destroyLocalSession();
            return false;
        }

        $_SESSION['last_activity_at'] = $now;
        return true;
    }

    public static function requireLogin(): void
    {
        if (!self::check()) redirect('login.php');
    }

    public static function user(): array
    {
        return is_array($_SESSION['user'] ?? null) ? $_SESSION['user'] : [];
    }

    public static function accessToken(): ?string
    {
        return isset($_SESSION['access_token']) ? (string) $_SESSION['access_token'] : null;
    }

    public static function refreshToken(): ?string
    {
        return isset($_SESSION['refresh_token']) ? (string) $_SESSION['refresh_token'] : null;
    }

    public static function api(ApiClient $api, string $method, string $path, ?array $body = null): array
    {
        self::requireLogin();

        try {
            return $api->request($method, $path, $body, self::accessToken());
        } catch (ApiException $exception) {
            if ($exception->status() !== 401 || !self::refreshToken()) {
                throw $exception;
            }

            try {
                $response = $api->request('POST', '/api/v1/auth/refresh', [
                    'refresh_token' => self::refreshToken(),
                ]);
            } catch (ApiException) {
                self::destroyLocalSession();
                throw new ApiException('Your session has expired. Please sign in again.', 401);
            }

            if (empty($response['access_token']) || empty($response['refresh_token'])) {
                self::destroyLocalSession();
                throw new ApiException('Your session has expired. Please sign in again.', 401);
            }

            $_SESSION['access_token'] = (string) $response['access_token'];
            $_SESSION['refresh_token'] = (string) $response['refresh_token'];
            $_SESSION['last_activity_at'] = time();

            return $api->request($method, $path, $body, self::accessToken());
        }
    }

    private static function destroyLocalSession(): void
    {
        $_SESSION = [];

        if (ini_get('session.use_cookies')) {
            $params = session_get_cookie_params();
            setcookie(session_name(), '', [
                'expires' => time() - 42000,
                'path' => $params['path'] ?? '/',
                'domain' => $params['domain'] ?? '',
                'secure' => (bool) ($params['secure'] ?? false),
                'httponly' => (bool) ($params['httponly'] ?? true),
                'samesite' => (string) ($params['samesite'] ?? 'Lax'),
            ]);
        }

        if (session_status() === PHP_SESSION_ACTIVE) {
            session_destroy();
        }
    }
}
