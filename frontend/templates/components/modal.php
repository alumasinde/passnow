<?php
$id ??= 'app-modal';
$title ??= '';
$size ??= 'md';
?>
<div class="modal-backdrop" data-modal="<?= e($id) ?>" hidden>
    <section class="modal modal-<?= e($size) ?>" role="dialog" aria-modal="true" aria-labelledby="<?= e($id) ?>-title">
        <header class="modal-header">
            <h2 id="<?= e($id) ?>-title"><?= e($title) ?></h2>
            <button type="button" class="icon-button modal-close" data-modal-close="<?= e($id) ?>" aria-label="Close">
                <i class="fa-solid fa-xmark"></i>
            </button>
        </header>
        <div class="modal-body">
            <?php if (!empty($slot)): require $slot; endif; ?>
        </div>
        <?php if (!empty($footer)): ?>
            <footer class="modal-footer"><?php require $footer; ?></footer>
        <?php endif; ?>
    </section>
</div>
