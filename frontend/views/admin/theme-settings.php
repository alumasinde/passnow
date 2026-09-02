<?php
$defaults = [
    'brand_name' => '',
    'logo_url' => '',
    'favicon_url' => '',
    'primary_color' => '#2563eb',
    'primary_color_dark' => '#1d4ed8',
    'accent_color' => '#2563eb',
    'sidebar_background' => '#ffffff',
    'sidebar_text' => '#475467',
    'sidebar_active_background' => '#eff6ff',
    'sidebar_active_text' => '#2563eb',
    'appearance' => 'light',
];
$data = array_merge($defaults, is_array($data ?? null) ? $data : []);
?>
<section class="page-header">
    <div>
        <span class="eyebrow">Tenant branding</span>
        <h1>Theme & branding</h1>
        <p>Configure your organization identity, colors, sidebar and appearance without hardcoding the frontend.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('settings')) ?>"><i class="fa-solid fa-arrow-left"></i> Settings</a>
</section>

<?php if ($errors): ?>
    <div class="alert alert-danger"><div><?php foreach ($errors as $error): ?><div><?= e($error) ?></div><?php endforeach; ?></div></div>
<?php endif; ?>

<form method="post" enctype="multipart/form-data" class="content-card form-card" data-loading-form>
    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
    <div class="form-grid theme-form-grid">
        <div class="theme-section-title field-full"><strong>Brand identity</strong><small>Used across the login page, sidebar, top bar and browser tab.</small></div>

        <div class="field">
            <label for="brand_name">Brand name</label>
            <input id="brand_name" name="brand_name" maxlength="160" value="<?= e((string)$data['brand_name']) ?>" placeholder="AlbaTech Solutions">
        </div>
        <div class="field">
            <label for="appearance">Appearance</label>
            <select id="appearance" name="appearance">
                <?php foreach (['light' => 'Light', 'dark' => 'Dark', 'system' => 'Follow device'] as $value => $label): ?>
                    <option value="<?= e($value) ?>" <?= $data['appearance'] === $value ? 'selected' : '' ?>><?= e($label) ?></option>
                <?php endforeach; ?>
            </select>
        </div>
        <div class="field">
            <label for="logo_url">Logo URL</label>
            <input id="logo_url" name="logo_url" value="<?= e((string)$data['logo_url']) ?>" placeholder="https://example.com/logo.png">
            <small class="field-help">HTTPS URL or root-relative path. You can also upload a logo below.</small>
        </div>
        <div class="field">
            <label for="favicon_url">Favicon URL</label>
            <input id="favicon_url" name="favicon_url" value="<?= e((string)$data['favicon_url']) ?>" placeholder="https://example.com/favicon.png">
        </div>
        <div class="field">
            <label for="logo_file">Upload logo</label>
            <input id="logo_file" name="logo_file" type="file" accept=".png,.jpg,.jpeg,.gif,.webp,.ico,image/png,image/jpeg,image/gif,image/webp,image/x-icon">
            <small class="field-help">PNG, JPEG, GIF, WebP or ICO. Maximum size is controlled by the tenant media service.</small>
        </div>
        <div class="field">
            <label for="favicon_file">Upload favicon</label>
            <input id="favicon_file" name="favicon_file" type="file" accept=".png,.jpg,.jpeg,.gif,.webp,.ico,image/png,image/jpeg,image/gif,image/webp,image/x-icon">
            <small class="field-help">Uploading a file replaces the corresponding URL when the theme is saved.</small>
        </div>
        <div class="field-full">
            <a class="btn btn-secondary" href="<?= e(url('media')) ?>"><i class="fa-solid fa-photo-film"></i> Open media library</a>
        </div>

        <div class="theme-section-title field-full"><strong>Application colors</strong><small>All colors are stored as tenant configuration and exposed through CSS variables.</small></div>

        <?php foreach ([
            'primary_color' => ['Primary color', 'Main buttons, active states and links'],
            'primary_color_dark' => ['Primary dark', 'Primary hover and emphasis states'],
            'accent_color' => ['Accent color', 'Highlights and semantic accents'],
            'sidebar_background' => ['Sidebar background', 'Tenant navigation background'],
            'sidebar_text' => ['Sidebar text', 'Default navigation text'],
            'sidebar_active_background' => ['Sidebar active background', 'Selected navigation item'],
            'sidebar_active_text' => ['Sidebar active text', 'Selected navigation item text'],
        ] as $key => [$label, $help]): ?>
            <div class="field theme-color-field">
                <label for="<?= e($key) ?>"><?= e($label) ?></label>
                <div class="color-input-wrap">
                    <input class="color-picker" type="color" id="<?= e($key) ?>_picker" value="<?= e((string)$data[$key]) ?>" data-theme-color-picker="<?= e($key) ?>">
                    <input id="<?= e($key) ?>" name="<?= e($key) ?>" maxlength="7" pattern="#[0-9A-Fa-f]{6}" value="<?= e((string)$data[$key]) ?>" data-theme-color-value="<?= e($key) ?>">
                </div>
                <small class="field-help"><?= e($help) ?></small>
            </div>
        <?php endforeach; ?>

        <div class="field-full theme-preview">
            <strong>Preview</strong>
            <div class="theme-preview-shell" id="theme-preview">
                <div class="theme-preview-sidebar">
                    <div class="theme-preview-brand">Your Brand</div>
                    <span class="active">Dashboard</span>
                    <span>Visitors</span>
                    <span>Gate Passes</span>
                </div>
                <div class="theme-preview-main">
                    <div class="theme-preview-card">
                        <small>Primary action</small>
                        <button type="button">Save changes</button>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div class="form-actions">
        <button class="btn btn-primary" type="submit" data-loading-label="Saving theme..."><span data-button-label>Save theme</span></button>
    </div>
</form>

<script>
document.addEventListener('DOMContentLoaded', function () {
    const preview = document.getElementById('theme-preview');
    const apply = () => {
        const values = {};
        document.querySelectorAll('[data-theme-color-value]').forEach((input) => values[input.dataset.themeColorValue] = input.value);
        preview.style.setProperty('--preview-primary', values.primary_color || '#2563eb');
        preview.style.setProperty('--preview-sidebar-bg', values.sidebar_background || '#ffffff');
        preview.style.setProperty('--preview-sidebar-text', values.sidebar_text || '#475467');
        preview.style.setProperty('--preview-active-bg', values.sidebar_active_background || '#eff6ff');
        preview.style.setProperty('--preview-active-text', values.sidebar_active_text || '#2563eb');
    };
    document.querySelectorAll('[data-theme-color-picker]').forEach((picker) => {
        const key = picker.dataset.themeColorPicker;
        const text = document.querySelector('[data-theme-color-value="' + key + '"]');
        picker.addEventListener('input', () => { text.value = picker.value; apply(); });
    });
    document.querySelectorAll('[data-theme-color-value]').forEach((input) => {
        input.addEventListener('input', () => {
            const picker = document.querySelector('[data-theme-color-picker="' + input.dataset.themeColorValue + '"]');
            if (/^#[0-9a-fA-F]{6}$/.test(input.value)) picker.value = input.value;
            apply();
        });
    });
    apply();
});
</script>