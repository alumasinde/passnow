<?php
declare(strict_types=1);

$navigation = [];
$navigationError = null;

try {
    $response = Auth::api(App::api(), 'GET', '/api/v1/navigation');
    $items = $response['items'] ?? [];
    if (is_array($items)) {
        $navigation = array_values(array_filter($items, static fn ($item): bool => is_array($item)));
    }
} catch (Throwable $e) {
    $navigationError = 'Navigation is temporarily unavailable.';
}

$currentPath = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';

$isActive = static function (array $item) use ($currentPath): bool {
    $prefixes = $item['match_prefixes'] ?? [];
    if (!is_array($prefixes)) {
        return false;
    }

    foreach ($prefixes as $prefix) {
        $prefix = '/' . ltrim((string)$prefix, '/');
        if ($prefix === '/') {
            continue;
        }

        if (
            $currentPath === $prefix ||
            str_starts_with($currentPath, $prefix . '/') ||
            str_ends_with($currentPath, $prefix) ||
            str_contains($currentPath, $prefix . '/')
        ) {
            return true;
        }
    }

    return false;
};

$mainNavigation = array_values(array_filter(
    $navigation,
    static fn (array $item): bool => ($item['placement'] ?? 'main') === 'main'
));

$bottomNavigation = array_values(array_filter(
    $navigation,
    static fn (array $item): bool => ($item['placement'] ?? 'main') === 'bottom'
));
?>
<aside class="sidebar" data-sidebar>
    <?php $theme = Theme::current(); ?>
    <div class="brand">
        <span class="brand-mark">
            <?php if (($theme['logo_url'] ?? '') !== ''): ?>
                <img src="<?= e((string)$theme['logo_url']) ?>" alt="<?= e(Theme::brandName()) ?> logo" class="brand-logo" data-tenant-logo>
            <?php else: ?>
                <i class="fa-solid fa-door-open"></i>
            <?php endif; ?>
        </span>
        <span class="brand-name" data-tenant-brand-name><?= e(Theme::brandName()) ?></span>
    </div>

    <nav class="sidebar-nav" aria-label="Main navigation">
        <?php foreach ($mainNavigation as $item): ?>
            <?php
            $label = (string)($item['label'] ?? '');
            $icon = (string)($item['icon'] ?? 'fa-circle');
            $href = (string)($item['href'] ?? '');
            if ($label === '' || $href === '') continue;
            ?>
            <a class="nav-item <?= $isActive($item) ? 'active' : '' ?>" href="<?= e(url($href)) ?>" title="<?= e($label) ?>">
                <i class="fa-solid <?= e($icon) ?>"></i>
                <span><?= e($label) ?></span>
            </a>
        <?php endforeach; ?>
    </nav>

    <div class="sidebar-bottom">
        <?php foreach ($bottomNavigation as $item): ?>
            <?php
            $label = (string)($item['label'] ?? '');
            $icon = (string)($item['icon'] ?? 'fa-circle');
            $href = (string)($item['href'] ?? '');
            if ($label === '' || $href === '') continue;
            ?>
            <a class="nav-item <?= $isActive($item) ? 'active' : '' ?>" href="<?= e(url($href)) ?>">
                <i class="fa-solid <?= e($icon) ?>"></i>
                <span><?= e($label) ?></span>
            </a>
        <?php endforeach; ?>

        <?php if ($navigationError !== null): ?>
            <span class="sidebar-status" title="<?= e($navigationError) ?>">
                <i class="fa-solid fa-triangle-exclamation"></i>
            </span>
        <?php endif; ?>
    </div>
</aside>
