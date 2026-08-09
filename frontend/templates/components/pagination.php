<?php
if (!isset($paginator) || !$paginator instanceof Paginator) return;
$query ??= new ListQuery();
?>
<?php if ($paginator->total() > 0): ?>
    <nav class="pagination" aria-label="Pagination">
        <div class="pagination-summary">
            Showing <?= e($paginator->from()) ?>–<?= e($paginator->to()) ?>
            of <?= e($paginator->total()) ?>
        </div>

        <div class="pagination-links">
            <?php if ($paginator->hasPrevious()): ?>
                <a class="page-link" href="?<?= e(http_build_query(array_merge($_GET, ['page' => $paginator->currentPage() - 1]))) ?>" aria-label="Previous">
                    <i class="fa-solid fa-chevron-left"></i>
                </a>
            <?php else: ?>
                <span class="page-link disabled"><i class="fa-solid fa-chevron-left"></i></span>
            <?php endif; ?>

            <?php foreach ($paginator->pages() as $page): ?>
                <?php if ($page === null): ?>
                    <span class="page-ellipsis">…</span>
                <?php elseif ((int) $page === $paginator->currentPage()): ?>
                    <span class="page-link active" aria-current="page"><?= e($page) ?></span>
                <?php else: ?>
                    <a class="page-link" href="?<?= e(http_build_query(array_merge($_GET, ['page' => $page]))) ?>">
                        <?= e($page) ?>
                    </a>
                <?php endif; ?>
            <?php endforeach; ?>

            <?php if ($paginator->hasNext()): ?>
                <a class="page-link" href="?<?= e(http_build_query(array_merge($_GET, ['page' => $paginator->currentPage() + 1]))) ?>" aria-label="Next">
                    <i class="fa-solid fa-chevron-right"></i>
                </a>
            <?php else: ?>
                <span class="page-link disabled"><i class="fa-solid fa-chevron-right"></i></span>
            <?php endif; ?>
        </div>
    </nav>
<?php endif; ?>
