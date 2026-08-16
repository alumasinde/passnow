<?php
declare(strict_types=1);

final class Csrf
{
    public static function token(): string
    {
        $now = time();
        $ttl = (int) AppContext::config('security.csrf_ttl', 3600);

        if (!isset($_SESSION['_csrf'], $_SESSION['_csrf_at']) || ($now - (int) $_SESSION['_csrf_at']) > $ttl) {
            self::rotate();
        }

        return (string) $_SESSION['_csrf'];
    }

    public static function rotate(): string
    {
        $_SESSION['_csrf'] = bin2hex(random_bytes(32));
        $_SESSION['_csrf_at'] = time();
        return (string) $_SESSION['_csrf'];
    }

    public static function verify(?string $token): bool
    {
        return is_string($token)
            && $token !== ''
            && !empty($_SESSION['_csrf'])
            && hash_equals((string) $_SESSION['_csrf'], $token)
            && (time() - (int) ($_SESSION['_csrf_at'] ?? 0)) <= (int) AppContext::config('security.csrf_ttl', 3600);
    }

    public static function requireValid(?string $token): void
    {
        if (self::verify($token)) return;

        http_response_code(419);
        if (class_exists('App')) {
            App::render('errors/419', ['title' => 'Security token expired'], 'guest');
            exit;
        }

        exit('The security token expired. Please try again.');
    }
}
