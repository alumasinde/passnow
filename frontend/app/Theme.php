<?php
declare(strict_types=1);

final class Theme
{
    private static ?array $theme = null;

    public static function current(): array
    {
        if (self::$theme !== null) {
            return self::$theme;
        }

        $defaults = [
            'brand_name' => '',
            'logo_url' => '',
            'favicon_url' => '',
            'primary_color' => '#2563eb',
            'primary_color_dark' => '#1d4ed8',
            'accent_color' => '#2563eb',
            'sidebar_background' => '#ffffff',
            'sidebar_text' => '#475467',
            'sidebar_active_background' => '#eff6ff',
            'sidebar_active_text' => '#2563eb',
            'appearance' => 'light',
        ];

        try {
            $data = App::api()->request('GET', '/api/v1/theme');
            if (is_array($data)) {
                foreach ($defaults as $key => $fallback) {
                    if (array_key_exists($key, $data) && is_scalar($data[$key])) {
                        $defaults[$key] = (string)$data[$key];
                    }
                }
            }
        } catch (Throwable) {
            // Branding must never prevent the application or login page from
            // rendering. Keep deterministic defaults if the API is unavailable.
        }

        self::$theme = self::sanitize($defaults);
        return self::$theme;
    }

    public static function forget(): void
    {
        self::$theme = null;
    }

    public static function brandName(): string
    {
        $name = trim((string)(self::current()['brand_name'] ?? ''));
        return $name !== '' ? $name : (string)App::config('app.name');
    }

    public static function bodyClass(): string
    {
        $appearance = (string)(self::current()['appearance'] ?? 'light');
        return $appearance === 'dark' ? 'theme-dark' : ($appearance === 'system' ? 'theme-system' : 'theme-light');
    }

    public static function faviconURL(): string
    {
        return (string)(self::current()['favicon_url'] ?? '');
    }

    public static function styleTag(): string
    {
        $t = self::current();
        $vars = [
            '--color-primary' => $t['primary_color'],
            '--color-primary-dark' => $t['primary_color_dark'],
            '--color-accent' => $t['accent_color'],
            '--sidebar-background' => $t['sidebar_background'],
            '--sidebar-text' => $t['sidebar_text'],
            '--sidebar-active-background' => $t['sidebar_active_background'],
            '--sidebar-active-text' => $t['sidebar_active_text'],
        ];
        $css = ':root{';
        foreach ($vars as $key => $value) {
            $css .= $key . ':' . $value . ';';
        }
        $css .= '}';

        if ($t['appearance'] === 'dark') {
            $css .= ':root{--color-text:#f9fafb;--color-muted:#98a2b3;--color-border:#344054;--color-surface:#101828;--color-background:#0b1220;--sidebar-background:#101828;--sidebar-text:#d0d5dd;--sidebar-active-background:#1d2939;}';
        } elseif ($t['appearance'] === 'system') {
            $css .= '@media (prefers-color-scheme: dark){:root{--color-text:#f9fafb;--color-muted:#98a2b3;--color-border:#344054;--color-surface:#101828;--color-background:#0b1220;--sidebar-background:#101828;--sidebar-text:#d0d5dd;--sidebar-active-background:#1d2939;}}';
        }

        return '<style id="tenant-theme">' . $css . '</style>';
    }

    private static function sanitize(array $theme): array
    {
        foreach (['primary_color','primary_color_dark','accent_color','sidebar_background','sidebar_text','sidebar_active_background','sidebar_active_text'] as $key) {
            if (!preg_match('/^#[0-9a-fA-F]{6}$/', (string)$theme[$key])) {
                $theme[$key] = match ($key) {
                    'primary_color' => '#2563eb',
                    'primary_color_dark' => '#1d4ed8',
                    'accent_color' => '#2563eb',
                    'sidebar_background' => '#ffffff',
                    'sidebar_text' => '#475467',
                    'sidebar_active_background' => '#eff6ff',
                    'sidebar_active_text' => '#2563eb',
                };
            }
        }
        if (!in_array($theme['appearance'], ['light','dark','system'], true)) {
            $theme['appearance'] = 'light';
        }
        foreach (['brand_name','logo_url','favicon_url'] as $key) {
            $theme[$key] = trim((string)$theme[$key]);
        }
        foreach (['logo_url','favicon_url'] as $key) {
            $theme[$key] = self::tenantAwareMediaURL((string)$theme[$key]);
        }
        return $theme;
    }

    private static function tenantAwareMediaURL(string $url): string
    {
        if ($url === '' || !str_contains($url, '/api/v1/media/public/')) return $url;
        $path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
        $segments = array_values(array_filter(explode('/', trim($path, '/'))));
        $reserved = ['login','logout','dashboard','platform','assets','api'];
        $slug = '';
        if ($segments !== [] && !in_array(strtolower($segments[0]), $reserved, true) && !str_ends_with(strtolower($segments[0]), '.php')) {
            $slug = strtolower($segments[0]);
        }
        if ($slug === '' || str_contains($url, '/'.$slug.'/api/v1/media/public/')) return $url;
        return str_replace('/api/v1/media/public/', '/'.$slug.'/api/v1/media/public/', $url);
    }
}
