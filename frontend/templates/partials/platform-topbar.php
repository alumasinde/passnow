<?php
declare(strict_types=1);

$user = is_array($_SESSION['platform_user'] ?? null) ? $_SESSION['platform_user'] : [];
$name = trim(($user['first_name'] ?? '') . ' ' . ($user['last_name'] ?? ''));
$initial = strtoupper(substr($name ?: 'A', 0, 1));
?>
<header class="topbar">
    <button class="icon-button mobile-menu" type="button" data-sidebar-toggle aria-label="Open menu">
        <i class="fa-solid fa-bars"></i>
    </button>
    <div class="topbar-title"><span class="muted">Platform Administration</span></div>
    <div class="topbar-actions">
        <div class="user-menu">
            <button class="user-trigger" type="button" data-user-menu aria-expanded="false">
                <span class="avatar"><?= e($initial) ?></span>
                <span class="user-name"><?= e($name ?: 'Platform Admin') ?></span>
                <i class="fa-solid fa-chevron-down"></i>
            </button>
            <div class="dropdown" data-user-dropdown hidden>
                <form method="post" action="<?= e(url('logout.php')) ?>">
                    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
                    <button type="submit"><i class="fa-solid fa-right-from-bracket"></i> Sign out</button>
                </form>
            </div>
        </div>
    </div>
</header>
