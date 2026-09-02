<?php
$user = Auth::user();
$name = trim(($user['first_name'] ?? '') . ' ' . ($user['last_name'] ?? ''));
$initial = strtoupper(substr($name ?: 'U', 0, 1));
?>
<header class="topbar">
    <button class="icon-button mobile-menu" type="button" data-sidebar-toggle aria-label="Open menu"><i class="fa-solid fa-bars"></i></button>
    <div class="topbar-title"><span class="muted"><?= e(Theme::brandName()) ?></span></div>
    <div class="topbar-actions"><div class="user-menu"><button class="user-trigger" type="button" data-user-menu aria-expanded="false"><span class="avatar"><?= e($initial) ?></span><span class="user-name"><?= e($name) ?></span><i class="fa-solid fa-chevron-down"></i></button><div class="dropdown" data-user-dropdown hidden><a href="<?= e(url('profile.php')) ?>"><i class="fa-regular fa-user"></i> Profile</a><form method="post" action="<?= e(url('logout')) ?>"><input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>"><button type="submit"><i class="fa-solid fa-right-from-bracket"></i> Sign out</button></form></div></div></div>
</header>