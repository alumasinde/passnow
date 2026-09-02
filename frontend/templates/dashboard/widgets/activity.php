<?php
$title=(string)($widget['title']??'Recent Activity');
$icon=(string)($widget['icon']??'clock');
$accent=preg_replace('/[^a-z]/','',(string)($widget['accent']??'primary'))?:'primary';
$items=is_array($widget['data']??null)?$widget['data']:[];
?>
<article class="dashboard-widget dashboard-widget-content dashboard-widget-activity dashboard-accent-<?= e($accent) ?>">
    <header class="dashboard-widget-header"><div><span class="dashboard-widget-title"><?= e($title) ?></span><small>Latest tenant activity</small></div><span class="dashboard-header-icon"><i class="fa-solid fa-<?= e($icon) ?>"></i></span></header>
    <?php if ($items === []): ?><div class="dashboard-empty">No recent activity yet.</div>
    <?php else: ?><div class="dashboard-activity-list"><?php foreach ($items as $item): ?>
        <div class="dashboard-activity-item"><span class="dashboard-activity-dot"></span><div><strong><?= e((string)($item['action']??'Activity')) ?></strong><small><?= e((string)($item['entity_type']??'System activity')) ?> · <?= e((string)($item['created_at']??'')) ?></small></div></div>
    <?php endforeach; ?></div><?php endif; ?>
</article>
