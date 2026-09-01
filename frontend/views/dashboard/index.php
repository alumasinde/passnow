<section class="page-header">
    <div><span class="eyebrow">Overview</span>
        <h1>Dashboard</h1>
        <p>Monitor your gatepass operations at a glance.</p>
    </div>
</section><?php if (!empty($error)): ?><div class="alert alert-warning"><i class="fa-solid fa-triangle-exclamation"></i><?= e($error) ?></div><?php endif; ?><div class="stats-grid"><?php $cards = [['key' => 'active_gatepasses', 'label' => 'Active gatepasses', 'icon' => 'fa-ticket'], ['key' => 'pending_approvals', 'label' => 'Pending approvals', 'icon' => 'fa-user-check'], ['key' => 'currently_on_premises', 'label' => 'On premises', 'icon' => 'fa-person-circle-check'], ['key' => 'overdue_gatepasses', 'label' => 'Overdue returns', 'icon' => 'fa-clock']];
                                                                                                                                                                                    foreach ($cards as $card): $value = $summary[$card['key']] ?? 0; ?><article class="stat-card">
            <div class="stat-icon"><i class="fa-solid <?= e($card['icon']) ?>"></i></div>
            <div><span class="stat-label"><?= e($card['label']) ?></span><strong class="stat-value"><?= e((string)$value) ?></strong></div>
        </article><?php endforeach; ?></div>
<section class="content-card">
    <div class="card-header">
        <div>
            <h2>Quick actions</h2>
            <p>Common operations for your workspace.</p>
        </div>
    </div>
    <div class="quick-actions"><a class="quick-action" href="<?= e(url('gatepasses.php')) ?>"><i class="fa-solid fa-right-left"></i><span>Manage gatepasses</span></a><a class="quick-action" href="<?= e(url('approvals.php')) ?>"><i class="fa-solid fa-user-check"></i><span>Review approvals</span></a><a class="quick-action" href="<?= e(url('visitors.php')) ?>"><i class="fa-solid fa-user-plus"></i><span>Manage visitors</span></a></div>
</section>