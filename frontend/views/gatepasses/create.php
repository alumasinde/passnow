<section class="page-header">
    <div>
        <span class="eyebrow">Gatepasses</span>
        <h1>New gatepass</h1>
        <p>Create a request using the rules configured for your tenant.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('gatepasses.php')) ?>">
        <i class="fa-solid fa-arrow-left"></i>
        Back
    </a>
</section>

<?php if ($errors): ?>
    <div class="alert alert-danger">
        <i class="fa-solid fa-circle-exclamation"></i>
        <div><?php foreach ($errors as $error): ?><div><?= e($error) ?></div><?php endforeach; ?></div>
    </div>
<?php endif; ?>

<form method="post" class="content-card form-card" data-loading-form>
    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">

    <div class="card-header">
        <h2>Gatepass details</h2>
        <p>Required fields are marked with an asterisk.</p>
    </div>

    <div class="form-grid">
        <?php component('select', [
            'name' => 'gatepass_type_id',
            'label' => 'Gatepass type',
            'options' => $types,
            'value' => oldOr('gatepass_type_id'),
            'required' => true,
            'placeholder' => 'Select gatepass type',
        ]); ?>

        <?php component('select', [
            'name' => 'subject_type',
            'label' => 'Person type',
            'options' => [
                ['value' => 'employee', 'label' => 'Employee'],
                ['value' => 'visitor', 'label' => 'Visitor'],
            ],
            'value' => oldOr('subject_type'),
            'required' => true,
            'placeholder' => 'Select person type',
        ]); ?>

        <div class="field">
            <label for="subject_id">Person <span class="required">*</span></label>
            <input id="subject_id" name="subject_id" type="number"
                   min="1" required value="<?= e((string)oldOr('subject_id')) ?>"
                   placeholder="Enter or select person ID">
            <small class="field-help">The person must already exist in this tenant.</small>
        </div>

        <?php component('select', [
            'name' => 'direction',
            'label' => 'Direction',
            'options' => [
                ['value' => 'out', 'label' => 'Going out'],
                ['value' => 'in', 'label' => 'Coming in'],
            ],
            'value' => oldOr('direction'),
            'required' => true,
            'placeholder' => 'Select direction',
        ]); ?>

        <div class="field field-full">
            <label for="purpose">Purpose <span class="required">*</span></label>
            <textarea id="purpose" name="purpose" maxlength="500" required
                      placeholder="Why is this movement required?"><?= e((string)oldOr('purpose')) ?></textarea>
        </div>

        <div class="field">
            <label class="checkbox-field">
                <input type="checkbox" name="returnable" value="1"
                    <?= oldOr('returnable') ? 'checked' : '' ?> data-returnable-toggle>
                <span>
                    <strong>Returnable</strong>
                    <small>This gatepass requires an eventual check-in.</small>
                </span>
            </label>
        </div>

        <div class="field" data-return-date>
            <label for="expected_return_at">Expected return</label>
            <input id="expected_return_at" name="expected_return_at" type="datetime-local"
                   value="<?= e((string)oldOr('expected_return_at')) ?>">
        </div>

        <div class="field field-full">
            <label for="notes">Notes</label>
            <textarea id="notes" name="notes" maxlength="1000"
                      placeholder="Optional notes"><?= e((string)oldOr('notes')) ?></textarea>
        </div>
    </div>

    <div class="form-actions">
        <a class="btn btn-secondary" href="<?= e(url('gatepasses.php')) ?>">Cancel</a>
        <button class="btn btn-primary" type="submit" data-loading-label="Creating...">
            <span data-button-label>Create gatepass</span>
        </button>
    </div>
</form>
