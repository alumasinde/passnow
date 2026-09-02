<?php
declare(strict_types=1);

$widgets = is_array($dashboard['widgets'] ?? null) ? $dashboard['widgets'] : [];
usort($widgets, static fn(array $a, array $b): int => ((int)($a['order'] ?? 0)) <=> ((int)($b['order'] ?? 0)));

$statWidgets = array_values(array_filter($widgets, static fn(array $w): bool => ($w['type'] ?? '') === 'stat'));
$contentWidgets = array_values(array_filter($widgets, static fn(array $w): bool => ($w['type'] ?? '') !== 'stat'));
?>
<div class="page-header">
    <div>
        <span class="eyebrow">Operations overview</span>
        <h1>Dashboard</h1>
        <p>Live operational information based on your current permissions.</p>
    </div>
</div>

<?php if ($error !== null): ?>
    <div class="alert alert-warning">
        <i class="fa-solid fa-triangle-exclamation"></i>
        <span><?= e($error) ?></span>
    </div>
<?php endif; ?>

<?php if ($statWidgets !== []): ?>
    <section class="dashboard-grid dashboard-stats-grid">
        <?php foreach ($statWidgets as $widget): ?>
            <?php $widgetFile = __DIR__ . '/widgets/stat.php'; if (is_file($widgetFile)) require $widgetFile; ?>
        <?php endforeach; ?>
    </section>
<?php endif; ?>

<?php if ($contentWidgets !== []): ?>
    <section class="dashboard-grid dashboard-content-grid">
        <?php foreach ($contentWidgets as $widget): ?>
            <?php
            $type = preg_replace('/[^a-z_]/', '', (string)($widget['type'] ?? ''));
            $widgetFile = __DIR__ . '/widgets/' . $type . '.php';
            if (is_file($widgetFile)) {
                require $widgetFile;
            }
            ?>
        <?php endforeach; ?>
    </section>
<?php elseif ($error === null): ?>
    <div class="content-card empty-state">
        <div class="empty-icon"><i class="fa-solid fa-chart-line"></i></div>
        <h3>No dashboard widgets available</h3>
        <p>Your current role does not have dashboard widgets assigned.</p>
    </div>
<?php endif; ?>
