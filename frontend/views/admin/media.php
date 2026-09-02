<section class="page-header">
    <div>
        <span class="eyebrow">Tenant media</span>
        <h1>Media library</h1>
        <p>Upload and manage tenant-owned images for branding and future modules.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('settings')) ?>" data-back><i class="fa-solid fa-arrow-left"></i> Settings</a>
</section>

<?php if ($errors): ?>
    <div class="alert alert-danger"><div><?php foreach ($errors as $error): ?><div><?= e($error) ?></div><?php endforeach; ?></div></div>
<?php endif; ?>

<section class="content-card form-card media-upload-card">
    <div class="card-header"><h2>Upload media</h2><p>PNG, JPEG, GIF, WebP or ICO files are accepted.</p></div>
    <form method="post" enctype="multipart/form-data" class="form-grid" data-loading-form>
        <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
        <input type="hidden" name="action" value="upload">
        <div class="field">
            <label for="file">Image file</label>
            <input id="file" name="file" type="file" required accept=".png,.jpg,.jpeg,.gif,.webp,.ico,image/png,image/jpeg,image/gif,image/webp,image/x-icon">
        </div>
        <div class="field">
            <label for="purpose">Purpose</label>
            <select id="purpose" name="purpose">
                <option value="general">General</option>
                <option value="branding">Branding</option>
            </select>
        </div>
        <div class="field-full form-actions">
            <button class="btn btn-primary" type="submit" data-loading-label="Uploading..."><span data-button-label>Upload file</span></button>
        </div>
    </form>
</section>

<section class="content-card media-library">
    <div class="card-header"><h2>Your files</h2><p>Files are isolated by tenant and stored through the PassNow media service.</p></div>
    <?php if (!$items): ?>
        <div class="empty-state"><i class="fa-regular fa-images"></i><strong>No media uploaded yet</strong><span>Upload your first logo, favicon or tenant image.</span></div>
    <?php else: ?>
        <div class="media-grid">
            <?php foreach ($items as $item): ?>
                <article class="media-card">
                    <a href="<?= e((string)($item['public_url'] ?? '')) ?>" target="_blank" rel="noopener" class="media-thumb">
                        <img src="<?= e((string)($item['public_url'] ?? '')) ?>" alt="<?= e((string)($item['original_name'] ?? 'Media file')) ?>">
                    </a>
                    <div class="media-meta">
                        <strong title="<?= e((string)($item['original_name'] ?? '')) ?>"><?= e((string)($item['original_name'] ?? 'Media file')) ?></strong>
                        <small><?= e((string)($item['mime_type'] ?? '')) ?> · <?= number_format(((int)($item['size_bytes'] ?? 0)) / 1024, 1) ?> KB</small>
                    </div>
                    <div class="media-actions">
                        <button type="button" class="btn btn-secondary btn-sm" data-copy-media="<?= e((string)($item['public_url'] ?? '')) ?>"><i class="fa-regular fa-copy"></i> Copy URL</button>
                        <form method="post" onsubmit="return confirm('Delete this media file?');">
                            <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
                            <input type="hidden" name="action" value="delete">
                            <input type="hidden" name="id" value="<?= (int)($item['id'] ?? 0) ?>">
                            <button type="submit" class="btn btn-danger btn-sm"><i class="fa-regular fa-trash-can"></i> Delete</button>
                        </form>
                    </div>
                </article>
            <?php endforeach; ?>
        </div>
    <?php endif; ?>
</section>

<script>
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-copy-media]').forEach(function (button) {
        button.addEventListener('click', async function () {
            const url = button.dataset.copyMedia || '';
            if (!url) return;
            try {
                await navigator.clipboard.writeText(url);
                const original = button.innerHTML;
                button.innerHTML = '<i class="fa-solid fa-check"></i> Copied';
                setTimeout(() => button.innerHTML = original, 1500);
            } catch (_) {
                window.prompt('Copy this URL:', url);
            }
        });
    });
});
</script>
