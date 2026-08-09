<section class="page-header">
    <div>
        <span class="eyebrow">Operations</span>
        <h1>Gatepasses</h1>
        <p>Search, filter and manage gatepass records.</p>
    </div>
    <a class="btn btn-primary" href="<?= e(url('gatepass-create.php')) ?>">
        <i class="fa-solid fa-plus"></i>
        New gatepass
    </a>
</section>

<?php if ($error): ?>
    <div class="alert alert-warning">
        <i class="fa-solid fa-triangle-exclamation"></i>
        <?= e($error) ?>
    </div>
<?php endif; ?>

<section class="content-card"><div class="toolbar-actions"><button type="button" class="btn btn-secondary" data-export-table="data-table"><i class="fa-solid fa-file-csv"></i> Export</button></div>
    <?php
    component('list-toolbar', [
        'query' => $query,
        'filters' => [
            ['name' => 'status', 'label' => 'Status', 'options' => $statusOptions],
            ['name' => 'type', 'label' => 'Type', 'options' => $typeOptions],
        ],
        'searchPlaceholder' => 'Search gatepass number or person...',
    ]);
    ?>

    <?php component('data-table', [
        'columns' => $columns,
        'rows' => $rows,
        'rowActions' => $rowActions,
        'emptyTitle' => 'No gatepasses found',
        'emptyMessage' => 'Try a different search or filter.',
    ]); ?>

    <?php component('pagination', ['paginator' => $paginator, 'query' => $query]); ?>
</section>
