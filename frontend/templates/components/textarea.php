<?php
$name ??= '';
$label ??= $name;
$value ??= '';
$required ??= false;
$maxlength ??= null;
$placeholder ??= '';
?>
<div class="field">
    <label for="<?= e($name) ?>">
        <?= e($label) ?><?php if ($required): ?> <span class="required">*</span><?php endif; ?>
    </label>
    <textarea id="<?= e($name) ?>" name="<?= e($name) ?>"
        <?= $required ? 'required' : '' ?>
        <?= $maxlength !== null ? 'maxlength="' . e((string)$maxlength) . '"' : '' ?>
        <?= $placeholder !== '' ? 'placeholder="' . e($placeholder) . '"' : '' ?>><?= e($value) ?></textarea>
</div>
