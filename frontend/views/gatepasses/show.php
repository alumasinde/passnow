<?php
$status = strtolower((string)($gatepass['status'] ?? ''));
$number = (string)($gatepass['gatepass_number'] ?? $gatepass['number'] ?? ('#' . ($gatepass['id'] ?? '')));
$canApprove = in_array($status, ['pending', 'awaiting_approval', 'pending_approval'], true);
$canCheckout = in_array($status, ['approved', 'ready', 'issued'], true);
$canCheckin = in_array($status, ['checked_out', 'out'], true);
?>
<section class="page-header">
    <div>
        <span class="eyebrow">Gatepass</span>
        <h1><?= e($number) ?></h1>
        <p><?= e((string)($gatepass['purpose'] ?? 'Gatepass details')) ?></p>
    </div>
    <a class="btn btn-secondary" data-back href="<?= e(url('gatepasses.php')) ?>" data-back>
        <i class="fa-solid fa-arrow-left"></i>
        Back
    </a>
</section>

<?php if ($error): ?>
    <div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><?= e($error) ?></div>
<?php endif; ?>

<?php if ($gatepass): ?>
<section class="detail-grid">
    <article class="content-card">
        <div class="card-header detail-header">
            <div>
                <h2>Gatepass information</h2>
                <p>Current state and configured movement details.</p>
            </div>
            <span class="status-badge status-<?= e(preg_replace('/[^a-z0-9_-]/i', '-', $status)) ?>">
                <?= e($gatepass['status'] ?? 'Unknown') ?>
            </span>
        </div>

        <dl class="detail-list">
            <?php
            $details = [
                'Gatepass number' => $number,
                'Type' => $gatepass['gatepass_type_name'] ?? $gatepass['type_name'] ?? $gatepass['type'] ?? '—',
                'Person' => $gatepass['subject_name'] ?? $gatepass['person_name'] ?? '—',
                'Direction' => $gatepass['direction'] ?? '—',
                'Returnable' => !empty($gatepass['returnable']) ? 'Yes' : 'No',
                'Expected return' => $gatepass['expected_return_at'] ?? '—',
                'Created' => $gatepass['created_at'] ?? '—',
            ];
            foreach ($details as $label => $value):
            ?>
                <div>
                    <dt><?= e($label) ?></dt>
                    <dd><?= e((string)$value) ?></dd>
                </div>
            <?php endforeach; ?>
        </dl>

        <?php if (!empty($gatepass['notes'])): ?>
            <div class="detail-note">
                <strong>Notes</strong>
                <p><?= e((string)$gatepass['notes']) ?></p>
            </div>
        <?php endif; ?>

        <div class="form-actions detail-actions">
            <?php if ($canApprove): ?>
                <a class="btn btn-primary" href="<?= e(url('gatepass-approvals.php?id=' . $id)) ?>">
                    <i class="fa-solid fa-user-check"></i> Review approval
                </a>
            <?php endif; ?>

            <?php if ($canCheckout): ?>
                <button class="btn btn-primary" type="button" data-modal-open="checkoutModal">
                    <i class="fa-solid fa-right-from-bracket"></i> Check out
                </button>
            <?php endif; ?>

            <?php if ($canCheckin): ?>
                <button class="btn btn-primary" type="button" data-modal-open="checkinModal">
                    <i class="fa-solid fa-right-to-bracket"></i> Check in
                </button>
            <?php endif; ?>

            <a class="btn btn-secondary" data-back href="<?= e(url('gatepass-qr.php?id=' . $id)) ?>" data-back>
                <i class="fa-solid fa-qrcode"></i> QR
            </a>
        </div>
    </article>

    <article class="content-card">
        <div class="card-header">
            <h2>Movement history</h2>
            <p>Recorded gatepass movements from the API.</p>
        </div>

        <?php component('data-table', [
            'columns' => [
                ['key' => 'movement_type', 'label' => 'Movement'],
                ['key' => 'status', 'label' => 'Status'],
                ['key' => 'performed_at', 'label' => 'Time'],
                ['key' => 'performed_by_name', 'label' => 'By'],
            ],
            'rows' => $movements,
            'emptyTitle' => 'No movements recorded',
            'emptyMessage' => 'Movement history will appear here after a gate operation.',
        ]); ?>
    </article>
</section>

<?php
$slot = __DIR__ . '/checkout-modal.php';
component('modal', ['id' => 'checkoutModal', 'title' => 'Check out gatepass', 'slot' => $slot, 'size' => 'sm']);
$slot = __DIR__ . '/checkin-modal.php';
component('modal', ['id' => 'checkinModal', 'title' => 'Check in gatepass', 'slot' => $slot, 'size' => 'sm']);
?>
<?php endif; ?>
