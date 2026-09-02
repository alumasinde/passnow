<?php
$title = (string)($widget['title'] ?? 'Overview');
$icon = (string)($widget['icon'] ?? 'chart-line');
$accent = preg_replace('/[^a-z]/', '', (string)($widget['accent'] ?? 'primary')) ?: 'primary';
$data = is_array($widget['data'] ?? null) ? $widget['data'] : [];
$series = is_array($data['series'] ?? null) ? $data['series'] : [];
$max = max([1, ...array_map(static fn($item): int => (int)($item['value'] ?? 0), $series)]);
?>
<article class="dashboard-widget dashboard-widget-content dashboard-widget-chart dashboard-accent-<?= e($accent) ?>">
    <header class="dashboard-widget-header">
        <div><span class="dashboard-widget-title"><?= e($title) ?></span><small>Current operational snapshot</small></div>
        <span class="dashboard-header-icon"><i class="fa-solid fa-<?= e($icon) ?>"></i></span>
    </header>
    <?php if ($series === []): ?>
        <div class="dashboard-empty">No chart data available yet.</div>
    <?php else: ?>
        <div class="dashboard-bars">
            <?php foreach ($series as $item): ?>
                <?php $value=(int)($item['value']??0); $height=max(8,(int)round(($value/$max)*100)); ?>
                <div class="dashboard-bar-item">
                    <div class="dashboard-bar-value"><?= e((string)$value) ?></div>
                    <div class="dashboard-bar-track"><span style="height: <?= $height ?>%"></span></div>
                    <div class="dashboard-bar-label"><?= e((string)($item['label']??$item['key']??'Item')) ?></div>
                </div>
            <?php endforeach; ?>
        </div>
    <?php endif; ?>
</article>
