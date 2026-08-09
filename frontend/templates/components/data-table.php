<?php
$columns ??= [];
$rows ??= [];
$rowActions ??= [];
$emptyTitle ??= 'No records found';
$emptyMessage ??= 'Try changing your search or filters.';
?>
<div class="table-wrap">
    <table id="data-table" class="data-table">
        <thead>
            <tr>
                <?php foreach ($columns as $column): ?>
                    <th><?= e((string) ($column['label'] ?? $column['key'] ?? '')) ?></th>
                <?php endforeach; ?>
                <?php if ($rowActions): ?><th class="actions-column">Actions</th><?php endif; ?>
            </tr>
        </thead>
        <tbody>
        <?php if (!$rows): ?>
            <tr>
                <td colspan="<?= e(count($columns) + ($rowActions ? 1 : 0)) ?>">
                    <?php component('empty-state', ['title' => $emptyTitle, 'message' => $emptyMessage]); ?>
                </td>
            </tr>
        <?php else: ?>
            <?php foreach ($rows as $row): ?>
                <tr>
                    <?php foreach ($columns as $column): ?>
                        <?php
                        $key = (string) ($column['key'] ?? '');
                        $value = $row[$key] ?? '';
                        ?>
                        <td data-label="<?= e((string) ($column['label'] ?? $key)) ?>">
                            <?php if (isset($column['render']) && is_callable($column['render'])): ?>
                                <?= $column['render']($value, $row) ?>
                            <?php else: ?>
                                <?= e($value) ?>
                            <?php endif; ?>
                        </td>
                    <?php endforeach; ?>

                    <?php if ($rowActions): ?>
                        <td class="table-actions">
                            <?php foreach ($rowActions as $action): ?>
                                <?php
                                $href = isset($action['href']) && is_callable($action['href'])
                                    ? $action['href']($row)
                                    : ($action['href'] ?? '#');
                                ?>
                                <?php if (($action['type'] ?? 'link') === 'button'): ?>
                                    <button type="button"
                                            class="btn btn-sm <?= e($action['class'] ?? 'btn-ghost') ?>"
                                            data-row-action="<?= e($action['name'] ?? '') ?>"
                                            data-id="<?= e($row[$action['id_key'] ?? 'id'] ?? '') ?>">
                                        <?php if (!empty($action['icon'])): ?><i class="fa-solid <?= e($action['icon']) ?>"></i><?php endif; ?>
                                        <?= e($action['label'] ?? 'Action') ?>
                                    </button>
                                <?php else: ?>
                                    <a class="btn btn-sm <?= e($action['class'] ?? 'btn-ghost') ?>"
                                       href="<?= e((string) $href) ?>">
                                        <?php if (!empty($action['icon'])): ?><i class="fa-solid <?= e($action['icon']) ?>"></i><?php endif; ?>
                                        <?= e($action['label'] ?? 'View') ?>
                                    </a>
                                <?php endif; ?>
                            <?php endforeach; ?>
                        </td>
                    <?php endif; ?>
                </tr>
            <?php endforeach; ?>
        <?php endif; ?>
        </tbody>
    </table>
</div>
