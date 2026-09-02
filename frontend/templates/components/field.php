<?php
$name ??= '';
$label ??= $name;
$type ??= 'text';
$required ??= false;
$maxlength ??= null;
$placeholder ??= '';
$error ??= '';
$autocomplete ??= '';
// Components must preserve the submitted value after validation failures.
// Explicit values (for edit forms) still win when the form has not been posted.
if (!array_key_exists('value', get_defined_vars()) || $value === null) {
    $value = $name !== '' && array_key_exists($name, $_POST) ? $_POST[$name] : '';
}
?>
<div class="field">
    <label for="<?= e($name) ?>">
        <?= e($label) ?><?php if ($required): ?> <span class="required">*</span><?php endif; ?>
    </label>
    <input
        id="<?= e($name) ?>"
        name="<?= e($name) ?>"
        type="<?= e($type) ?>"
        value="<?= e((string)$value) ?>"
        <?= $required ? 'required' : '' ?>
        <?= $maxlength !== null ? 'maxlength="' . e((string)$maxlength) . '"' : '' ?>
        <?= $placeholder !== '' ? 'placeholder="' . e($placeholder) . '"' : '' ?>
        <?= $autocomplete !== '' ? 'autocomplete="' . e($autocomplete) . '"' : '' ?>>
    <?php if ($error): ?><div class="field-error"><?= e($error) ?></div><?php endif; ?>
</div>
