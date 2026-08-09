<form method="post" action="<?= e(url('gatepass-operation.php')) ?>" data-loading-form>
    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
    <input type="hidden" name="id" value="<?= e((string)$id) ?>">
    <input type="hidden" name="operation" value="check-out">
    <p>Confirm that the person and items have been physically verified at the gate.</p>
    <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-modal-close="checkoutModal">Cancel</button>
        <button type="submit" class="btn btn-primary" data-loading-label="Checking out...">
            <span data-button-label>Confirm check-out</span>
        </button>
    </div>
</form>
