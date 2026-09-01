<?php $title = 'Tenant details'; ?><section class="page-header">
    <div><span class="eyebrow">Platform</span>
        <h1><?= e($tenant['name'] ?? 'Tenant') ?></h1>
        <p>Organization details, status and access domains.</p>
    </div><a class="btn btn-secondary" href="<?= e(url('platform/tenants')) ?>">Back</a>
</section>
<?php if ($error): ?><div class="alert alert-danger"><?= e($error) ?></div><?php endif ?>
<?php if ($tenant): ?><section class="content-card">
        <div class="form-grid">
            <div><strong>Status</strong>
                <p><?= e((string)$tenant['status']) ?></p>
            </div>
            <div><strong>Slug</strong>
                <p><?= e((string)$tenant['slug']) ?></p>
            </div>
            <div><strong>Created</strong>
                <p><?= e((string)$tenant['created_at']) ?></p>
            </div>
        </div>
    </section>
    <section class="content-card">
        <div class="page-header">
            <div>
                <h2>Access domains</h2>
                <p class="muted">Domains that identify this tenant.</p>
            </div><a class="btn btn-primary" href="<?= e(url('platform/tenant-edit?id=' . (int)$tenant['id'])) ?>">Edit tenant</a>
        </div>
        <div class="table-wrap">
            <table class="data-table">
                <thead>
                    <tr>
                        <th>Domain</th>
                        <th>Type</th>
                        <th>Primary</th>
                        <th>Verified</th>
                    </tr>
                </thead>
                <tbody><?php foreach (($tenant['domains'] ?? []) as $d): ?><tr>
                            <td><?= e($d['domain']) ?></td>
                            <td><?= e($d['type']) ?></td>
                            <td><?= !empty($d['is_primary']) ? 'Yes' : 'No' ?></td>
                            <td><?= !empty($d['is_verified']) ? 'Yes' : 'Pending' ?></td>
                        </tr><?php endforeach ?></tbody>
            </table>
        </div>
    </section><?php endif ?>