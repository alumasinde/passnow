<form method="post" action="<?= e(url('gatepass-operation.php')) ?>" data-loading-form>
    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
    <input type="hidden" name="id" value="<?= e((string)$id) ?>">
    <input type="hidden" name="operation" value="check-in">
    <p>Confirm that the person and any returnable items have been physically verified back at the gate.</p>
    <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-modal-close="checkinModal">Cancel</button>
        <button type="submit" class="btn btn-primary" data-loading-label="Checking in...">
            <span data-button-label>Confirm check-in</span>
        </button>
    </div>
</form>
