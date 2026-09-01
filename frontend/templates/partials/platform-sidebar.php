<?php
declare(strict_types=1);

$current = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
$isAdmin = in_array(rtrim($current, '/') ?: '/', ['/platform', '/platform.php', '/platform/tenants'], true);
?>
<aside class="sidebar" data-sidebar>
    <div class="brand">
        <span class="brand-mark"><i class="fa-solid fa-door-open"></i></span>
        <span class="brand-name"><?= e(App::config('app.name')) ?></span>
    </div>

    <nav class="sidebar-nav" aria-label="Platform administration">
        <div class="sidebar-section-label">PLATFORM</div>
        <a class="nav-item <?= $isAdmin ? 'active' : '' ?>" href="<?= e(url('platform/tenants')) ?>">
            <i class="fa-solid fa-shield-halved"></i>
            <span>Admin</span>
        </a>
    </nav>

    <div class="sidebar-bottom">
        <form method="post" action="<?= e(url('logout.php')) ?>">
            <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
            <button type="submit" class="nav-item nav-item-button">
                <i class="fa-solid fa-right-from-bracket"></i>
                <span>Sign out</span>
            </button>
        </form>
    </div>
</aside>
