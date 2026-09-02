<?php
$title = (string)($widget['title'] ?? 'Quick Actions');
$icon = (string)($widget['icon'] ?? 'bolt');
$accent = preg_replace('/[^a-z]/', '', (string)($widget['accent'] ?? 'primary')) ?: 'primary';
$actions = is_array($widget['data'] ?? null) ? $widget['data'] : [];
?>
<article class="dashboard-widget dashboard-widget-content dashboard-widget-actions dashboard-accent-<?= e($accent) ?>">
    <header class="dashboard-widget-header">
        <div>
            <span class="dashboard-widget-title"><?= e($title) ?></span>
            <small>Actions available to your current role</small>
        </div>
        <span class="dashboard-header-icon"><i class="fa-solid fa-<?= e($icon) ?>"></i></span>
    </header>
    <?php if ($actions === []): ?>
        <div class="dashboard-empty">No actions available.</div>
    <?php else: ?>
        <div class="dashboard-action-list">
            <?php foreach ($actions as $action): ?>
                <?php
                $label = (string)($action['label'] ?? 'Open');
                $actionIcon = (string)($action['icon'] ?? 'arrow-right');
                $route = (string)($action['route'] ?? '');
                if ($route === '') continue;
                ?>
                <a class="dashboard-action" href="<?= e(url($route)) ?>">
                    <span class="dashboard-action-icon"><i class="fa-solid fa-<?= e($actionIcon) ?>"></i></span>
                    <span><?= e($label) ?></span>
                    <i class="fa-solid fa-chevron-right dashboard-action-chevron"></i>
                </a>
            <?php endforeach; ?>
        </div>
    <?php endif; ?>
</article>