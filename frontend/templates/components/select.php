<?php
$name ??= '';
$label ??= $name;
$options ??= [];
$required ??= false;
$placeholder ??= 'Select...';
if (!array_key_exists('value', get_defined_vars()) || $value === null) {
    $value = $name !== '' && array_key_exists($name, $_POST) ? $_POST[$name] : '';
}
?>
<div class="field">
    <label for="<?= e($name) ?>">
        <?= e($label) ?><?php if ($required): ?> <span class="required">*</span><?php endif; ?>
    </label>
    <select id="<?= e($name) ?>" name="<?= e($name) ?>" <?= $required ? 'required' : '' ?>>
        <?php if (!$required): ?><option value=""><?= e($placeholder) ?></option><?php endif; ?>
        <?php foreach ($options as $option): ?>
            <?php
            $ov = is_array($option) ? (string)($option['value'] ?? $option['id'] ?? '') : (string)$option;
            $ol = is_array($option) ? (string)($option['label'] ?? $option['name'] ?? $ov) : (string)$option;
            ?>
            <option value="<?= e($ov) ?>" <?= (string)$value === $ov ? 'selected' : '' ?>><?= e($ol) ?></option>
        <?php endforeach; ?>
    </select>
</div>
