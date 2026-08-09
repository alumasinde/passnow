<?php
$name ??= '';
$label ??= $name;
$type ??= 'text';
$value ??= '';
$required ??= false;
$maxlength ??= null;
$placeholder ??= '';
$error ??= '';
?>
<div class="field">
    <label for="<?= e($name) ?>">
        <?= e($label) ?><?php if ($required): ?> <span class="required">*</span><?php endif; ?>
    </label>
    <input
        id="<?= e($name) ?>"
        name="<?= e($name) ?>"
        type="<?= e($type) ?>"
        value="<?= e($value) ?>"
        <?= $required ? 'required' : '' ?>
        <?= $maxlength !== null ? 'maxlength="' . e((string)$maxlength) . '"' : '' ?>
        <?= $placeholder !== '' ? 'placeholder="' . e($placeholder) . '"' : '' ?>>
    <?php if ($error): ?><div class="field-error"><?= e($error) ?></div><?php endif; ?>
</div>
