<?php
$title = 'Tenant database';
$isHealthy = !empty($health['healthy']);
$status = strtolower((string)($health['status'] ?? 'unknown'));
$statusLabel = $isHealthy ? 'Healthy' : ($status !== '' ? ucfirst($status) : 'Unknown');
$database = (string)($health['database'] ?? 'Unavailable');
$message = (string)($health['message'] ?? 'Health check not available.');
?>
<section class="tenant-db-page">
    <section class="tenant-db-header">
        <div class="tenant-db-heading">
            <a class="tenant-db-back-link" href="<?= e(url('platform/tenant-view?id='.(int)$id)) ?>">
                <i class="fa-solid fa-arrow-left"></i><span>Back to tenant</span>
            </a>
            <span class="eyebrow">Platform · Database Operations</span>
            <h1><?= e($tenant['name'] ?? 'Tenant') ?> database</h1>
            <p>Monitor the tenant database connection and safely apply pending tenant migrations.</p>
        </div>
        <div class="tenant-db-actions">
            <a class="btn btn-secondary" href="<?= e(url('platform/tenant-database?id='.(int)$id)) ?>">
                <i class="fa-solid fa-rotate-right"></i> Refresh health
            </a>
            <?php if($isHealthy): ?>
                <a class="btn btn-primary" href="#tenant-migrations">
                    <i class="fa-solid fa-database"></i> Run migrations
                </a>
            <?php endif; ?>
        </div>
    </section>

    <?php if($error): ?>
        <div class="alert alert-danger"><?= e($error) ?></div>
    <?php endif ?>

    <section class="tenant-db-card tenant-health-card <?= $isHealthy ? 'is-healthy' : 'is-unhealthy' ?>">
        <div class="tenant-db-card-header">
            <div class="tenant-db-title-group">
                <div class="tenant-db-icon health-icon">
                    <i class="fa-solid <?= $isHealthy ? 'fa-shield-heart' : 'fa-triangle-exclamation' ?>"></i>
                </div>
                <div>
                    <h2>Database health</h2>
                    <p>Live connection check against this tenant's isolated MySQL database.</p>
                </div>
            </div>
            <span class="tenant-health-badge <?= $isHealthy ? 'healthy' : 'unhealthy' ?>">
                <i class="fa-solid <?= $isHealthy ? 'fa-circle-check' : 'fa-circle-xmark' ?>"></i>
                <?= e($statusLabel) ?>
            </span>
        </div>

        <div class="tenant-health-details">
            <div class="tenant-health-detail">
                <span class="tenant-health-detail-icon"><i class="fa-solid fa-database"></i></span>
                <div><span class="detail-label">Database</span><strong><?= e($database) ?></strong></div>
            </div>
            <div class="tenant-health-detail">
                <span class="tenant-health-detail-icon"><i class="fa-solid fa-heart-pulse"></i></span>
                <div><span class="detail-label">Status</span><strong class="<?= $isHealthy ? 'text-success' : 'text-danger' ?>"><?= e($statusLabel) ?></strong></div>
            </div>
        </div>

        <div class="tenant-health-message <?= $isHealthy ? 'success' : 'danger' ?>">
            <span><i class="fa-solid <?= $isHealthy ? 'fa-circle-check' : 'fa-triangle-exclamation' ?>"></i></span>
            <div>
                <strong><?= $isHealthy ? 'Database connection is healthy' : 'Database connection needs attention' ?></strong>
                <p><?= e($message) ?></p>
            </div>
        </div>
    </section>

    <section id="tenant-migrations" class="tenant-db-card tenant-migrations-card">
        <div class="tenant-db-card-header">
            <div class="tenant-db-title-group">
                <div class="tenant-db-icon migration-icon"><i class="fa-solid fa-layer-group"></i></div>
                <div>
                    <h2>Tenant migrations</h2>
                    <p>Apply pending schema changes to this tenant database.</p>
                </div>
            </div>
        </div>

        <div class="tenant-migration-body">
            <div class="tenant-migration-info">
                <div class="tenant-migration-copy">
                    <h3>Safe tenant-only migrations</h3>
                    <p>Only the repository's <code>migrations/tenant</code> set is executed against <strong><?= e($database) ?></strong>. Platform migrations and other tenant databases are not touched.</p>
                </div>
                <div class="tenant-migration-note">
                    <i class="fa-solid fa-circle-info"></i>
                    <div><strong>Safe to run repeatedly</strong><span>Already applied migrations are skipped. Only pending migrations are executed.</span></div>
                </div>
            </div>

            <div class="tenant-migration-action">
                <?php if($isHealthy): ?>
                    <form method="post" data-loading-form>
                        <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
                        <input type="hidden" name="action" value="migrate">
                        <button class="btn btn-primary tenant-run-migrations" data-confirm="Run pending tenant migrations for <?= e($tenant['name'] ?? 'this tenant') ?>?">
                            <i class="fa-solid fa-database"></i><span>Run migrations</span>
                        </button>
                    </form>
                    <span class="tenant-action-hint"><i class="fa-solid fa-lock"></i> Runs only for this tenant</span>
                <?php else: ?>
                    <div class="tenant-migrations-blocked">
                        <i class="fa-solid fa-lock"></i>
                        <div><strong>Migrations are unavailable</strong><span>Restore database health before running migrations.</span></div>
                    </div>
                <?php endif; ?>
            </div>
        </div>
    </section>
</section>
