<section class="page-header">
    <div>
        <span class="eyebrow">Gatepass</span>
        <h1>QR code</h1>
        <p>Use the backend-generated QR endpoint for this gatepass.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('gatepass.php?id=' . (string) $id)) ?>" data-back>Back</a>
</section>

<section class="content-card qr-card">
    <?php if (!empty($error)): ?>
        <div class="alert alert-danger">
            <i class="fa-solid fa-circle-exclamation"></i>
            <?= e((string) $error) ?>
        </div>
    <?php elseif (!empty($token)): ?>
        <?php
        $qrURL = url(
            'gatepass-qr-image.php?token='
            . rawurlencode((string) $token)
        );
        ?>
        <img
            src="<?= e($qrURL) ?>"
            alt="Gatepass QR code"
            class="qr-image"
        >
        <p class="muted">The QR token is intentionally not displayed as plain text.</p>
    <?php else: ?>
        <?php component('empty-state', [
            'title' => 'QR unavailable',
            'message' => 'The API did not return a QR token for this gatepass.',
        ]); ?>
    <?php endif; ?>
</section>
