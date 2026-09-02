<?php
$title = (string)($widget['title'] ?? 'Metric');
$icon = (string)($widget['icon'] ?? 'circle');
$accent = preg_replace('/[^a-z]/', '', (string)($widget['accent'] ?? 'primary')) ?: 'primary';
$value = $widget['value'] ?? 0;
?>
<article class="dashboard-widget dashboard-widget-stat dashboard-accent-<?= e($accent) ?>">
    <div class="dashboard-widget-icon"><i class="fa-solid fa-<?= e($icon) ?>"></i></div>
    <div>
        <span class="dashboard-widget-label"><?= e($title) ?></span>
        <strong class="dashboard-widget-value"><?= e(is_scalar($value) ? (string)$value : '—') ?></strong>
    </div>
</article>
