<?php
$name ??= '';
$label ??= $name;
$options ??= [];
$value ??= '';
$required ??= false;
$placeholder ??= 'Select...';
?>
<div class="field">
    <label for="<?= e($name) ?>">
        <?= e($label) ?><?php if ($required): ?> <span class="required">*</span><?php endif; ?>
    </label>
    <select id="<?= e($name) ?>" name="<?= e($name) ?>" <?= $required ? 'required' : '' ?>>
        <?php if (!$required): ?><option value=""><?= e($placeholder) ?></option><?php endif; ?>
        <?php foreach ($options as $option): ?>
            <?php
            $ov = is_array($option) ? (string)($option['value'] ?? '') : (string)$option;
            $ol = is_array($option) ? (string)($option['label'] ?? $ov) : (string)$option;
            ?>
            <option value="<?= e($ov) ?>" <?= (string)$value === $ov ? 'selected' : '' ?>><?= e($ol) ?></option>
        <?php endforeach; ?>
    </select>
</div>
