<?php
$query ??= new ListQuery();
$filters ??= [];
$showSearch ??= true;
$showPerPage ??= true;
?>
<div class="list-toolbar">
    <form method="get" class="list-search-form">
        <?php foreach ($_GET as $key => $value): ?>
            <?php if ($key === 'q' || $key === 'page' || $key === 'per_page' || !is_string($value)) continue; ?>
            <input type="hidden" name="<?= e($key) ?>" value="<?= e($value) ?>">
        <?php endforeach; ?>

        <?php if ($showSearch): ?>
            <div class="search-box">
                <i class="fa-solid fa-magnifying-glass"></i>
                <input
                    type="search"
                    name="q"
                    value="<?= e($query->search()) ?>"
                    placeholder="<?= e($searchPlaceholder ?? 'Search...') ?>"
                    maxlength="100"
                    autocomplete="off">
            </div>
        <?php endif; ?>

        <?php foreach ($filters as $filter): ?>
            <?php
            $name = (string) ($filter['name'] ?? '');
            $label = (string) ($filter['label'] ?? '');
            $options = is_array($filter['options'] ?? null) ? $filter['options'] : [];
            $selected = (string) ($_GET[$name] ?? '');
            ?>
            <?php if ($name === '') continue; ?>
            <label class="filter-control">
                <span class="sr-only"><?= e($label ?: $name) ?></span>
                <select name="<?= e($name) ?>">
                    <option value=""><?= e($label ?: 'All') ?></option>
                    <?php foreach ($options as $option): ?>
                        <?php
                        $optionValue = is_array($option) ? (string) ($option['value'] ?? '') : (string) $option;
                        $optionLabel = is_array($option) ? (string) ($option['label'] ?? $optionValue) : (string) $option;
                        ?>
                        <option value="<?= e($optionValue) ?>" <?= $selected === $optionValue ? 'selected' : '' ?>>
                            <?= e($optionLabel) ?>
                        </option>
                    <?php endforeach; ?>
                </select>
            </label>
        <?php endforeach; ?>

        <button class="btn btn-secondary" type="submit">
            <i class="fa-solid fa-filter"></i>
            <span>Apply</span>
        </button>

        <?php if ($query->search() !== '' || count($filters) > 0 && count(array_intersect_key($_GET, array_flip(array_map(
            static fn($f) => (string) ($f['name'] ?? ''), $filters
        )))) > 0): ?>
            <a class="btn btn-ghost" href="<?= e(basename($_SERVER['PHP_SELF'] ?? '')) ?>">
                Clear
            </a>
        <?php endif; ?>
    </form>

    <?php if ($showPerPage): ?>
        <form method="get" class="per-page-form">
            <?php foreach ($_GET as $key => $value): ?>
                <?php if ($key === 'per_page' || $key === 'page' || !is_string($value)) continue; ?>
                <input type="hidden" name="<?= e($key) ?>" value="<?= e($value) ?>">
            <?php endforeach; ?>
            <label>
                <span class="muted">Rows</span>
                <select name="per_page" onchange="this.form.submit()">
                    <?php foreach ([10, 20, 50, 100] as $size): ?>
                        <option value="<?= $size ?>" <?= $query->perPage((int) App::config('ui.page_size', 20)) === $size ? 'selected' : '' ?>>
                            <?= $size ?>
                        </option>
                    <?php endforeach; ?>
                </select>
            </label>
        </form>
    <?php endif; ?>
</div>
