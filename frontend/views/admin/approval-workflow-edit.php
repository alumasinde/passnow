<section class="page-header">
    <div>
        <span class="eyebrow">Administration</span>
        <h1><?= $id ? 'Edit' : 'New' ?> approval workflow</h1>
        <p>Build the approval chain step by step. No JSON configuration is required.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('approval-workflows.php')) ?>" data-back>Back</a>
</section>

<?php if ($errors): ?>
<div class="alert alert-danger"><?php foreach ($errors as $x): ?><div><?= e($x) ?></div><?php endforeach; ?></div>
<?php endif; ?>

<form method="post" class="content-card form-card" data-loading-form>
    <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">

    <div class="form-grid">
        <?php component('field', ['name'=>'name','label'=>'Workflow name','value'=>(string)($item['name'] ?? ''),'required'=>true]); ?>
        <div class="field">
            <label class="checkbox-row">
                <input type="checkbox" name="active" value="1" <?= !array_key_exists('active',$item) || !empty($item['active']) ? 'checked' : '' ?>>
                <span>Workflow is active</span>
            </label>
            <small class="field-help">Inactive workflows remain available for history but cannot be selected for new gatepass types.</small>
        </div>
    </div>

    <div class="workflow-builder" data-workflow-builder>
        <div class="workflow-builder-header">
            <div>
                <h2>Approval steps</h2>
                <p class="muted">Approvals run from top to bottom. Reorder steps before saving.</p>
            </div>
            <button type="button" class="btn btn-secondary" data-add-step><i class="fa-solid fa-plus"></i> Add step</button>
        </div>

        <div data-steps>
        <?php foreach ($steps as $index => $step):
            $type = (string)($step['approver_type'] ?? 'role');
            $selected = $type === 'specific_user' ? (int)($step['user_id'] ?? 0) : (int)($step['role_id'] ?? 0);
        ?>
            <article class="workflow-step" data-step>
                <div class="workflow-step-number" data-step-number><?= $index + 1 ?></div>
                <div class="workflow-step-body">
                    <div class="form-grid">
                        <div class="field">
                            <label>Step name</label>
                            <input type="text" name="steps[<?= $index ?>][label]" value="<?= e((string)($step['label'] ?? '')) ?>" placeholder="e.g. Department Manager" required>
                        </div>
                        <div class="field">
                            <label>Approver type</label>
                            <select name="steps[<?= $index ?>][approver_type]" data-approver-type>
                                <option value="role" <?= $type === 'role' ? 'selected' : '' ?>>Role</option>
                                <option value="specific_user" <?= $type === 'specific_user' ? 'selected' : '' ?>>Specific user</option>
                            </select>
                        </div>
                        <div class="field">
                            <label>Approver</label>
                            <select name="steps[<?= $index ?>][approver_id]" data-approver-select required>
                                <?php if ($type === 'specific_user'): ?>
                                    <option value="">Select user</option>
                                    <?php foreach ($users as $user): ?>
                                        <option value="<?= (int)($user['id'] ?? 0) ?>" <?= $selected === (int)($user['id'] ?? 0) ? 'selected' : '' ?>><?= e((string)($user['name'] ?? $user['email'] ?? ('User #' . ($user['id'] ?? '')))) ?></option>
                                    <?php endforeach; ?>
                                <?php else: ?>
                                    <option value="">Select role</option>
                                    <?php foreach ($roles as $role): ?>
                                        <option value="<?= (int)($role['id'] ?? 0) ?>" <?= $selected === (int)($role['id'] ?? 0) ? 'selected' : '' ?>><?= e((string)($role['name'] ?? ('Role #' . ($role['id'] ?? '')))) ?></option>
                                    <?php endforeach; ?>
                                <?php endif; ?>
                            </select>
                        </div>
                        <div class="field">
                            <label class="checkbox-row">
                                <input type="checkbox" name="steps[<?= $index ?>][required]" value="1" <?= !array_key_exists('required',$step) || !empty($step['required']) ? 'checked' : '' ?>>
                                <span>Required approval</span>
                            </label>
                        </div>
                    </div>
                </div>
                <div class="workflow-step-actions">
                    <button type="button" class="btn btn-secondary" data-move-up title="Move up"><i class="fa-solid fa-arrow-up"></i></button>
                    <button type="button" class="btn btn-secondary" data-move-down title="Move down"><i class="fa-solid fa-arrow-down"></i></button>
                    <button type="button" class="btn btn-danger" data-remove-step title="Remove"><i class="fa-solid fa-trash"></i></button>
                </div>
            </article>
        <?php endforeach; ?>
        </div>
    </div>

    <div class="form-actions">
        <a class="btn btn-secondary" href="<?= e(url('approval-workflows.php')) ?>">Cancel</a>
        <button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save workflow</span></button>
    </div>
</form>

<template id="workflow-step-template">
    <article class="workflow-step" data-step>
        <div class="workflow-step-number" data-step-number></div>
        <div class="workflow-step-body">
            <div class="form-grid">
                <div class="field"><label>Step name</label><input type="text" data-name="label" placeholder="e.g. Department Manager" required></div>
                <div class="field"><label>Approver type</label><select data-name="approver_type" data-approver-type><option value="role">Role</option><option value="specific_user">Specific user</option></select></div>
                <div class="field"><label>Approver</label><select data-name="approver_id" data-approver-select required><option value="">Select role</option><?php foreach ($roles as $role): ?><option value="<?= (int)($role['id'] ?? 0) ?>"><?= e((string)($role['name'] ?? '')) ?></option><?php endforeach; ?></select></div>
                <div class="field"><label class="checkbox-row"><input type="checkbox" data-name="required" value="1" checked><span>Required approval</span></label></div>
            </div>
        </div>
        <div class="workflow-step-actions">
            <button type="button" class="btn btn-secondary" data-move-up><i class="fa-solid fa-arrow-up"></i></button>
            <button type="button" class="btn btn-secondary" data-move-down><i class="fa-solid fa-arrow-down"></i></button>
            <button type="button" class="btn btn-danger" data-remove-step><i class="fa-solid fa-trash"></i></button>
        </div>
    </article>
</template>

<script>
(() => {
    const builder = document.querySelector('[data-workflow-builder]');
    if (!builder) return;
    const stepsBox = builder.querySelector('[data-steps]');
    const template = document.getElementById('workflow-step-template');
    const roles = <?= json_encode(array_map(static fn($r) => ['id'=>(int)($r['id']??0),'name'=>(string)($r['name']??'')], $roles), JSON_UNESCAPED_SLASHES) ?>;
    const users = <?= json_encode(array_map(static fn($u) => ['id'=>(int)($u['id']??0),'name'=>(string)($u['name']??$u['email']??'')], $users), JSON_UNESCAPED_SLASHES) ?>;

    const optionsFor = type => {
        const rows = type === 'specific_user' ? users : roles;
        const first = type === 'specific_user' ? 'Select user' : 'Select role';
        return '<option value="">' + first + '</option>' + rows.map(row => '<option value="' + row.id + '">' + String(row.name).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#039;'}[c])) + '</option>').join('');
    };

    const renumber = () => {
        [...stepsBox.querySelectorAll('[data-step]')].forEach((step, index) => {
            step.querySelector('[data-step-number]').textContent = index + 1;
            step.querySelectorAll('[data-name]').forEach(el => el.name = 'steps[' + index + '][' + el.dataset.name + ']');
            step.querySelectorAll('input[name],select[name]').forEach(el => {
                if (!el.dataset.name) {
                    el.name = el.name.replace(/steps\[\d+\]/, 'steps[' + index + ']');
                }
            });
        });
    };

    builder.querySelector('[data-add-step]').addEventListener('click', () => {
        stepsBox.appendChild(template.content.cloneNode(true));
        renumber();
    });

    stepsBox.addEventListener('click', event => {
        const step = event.target.closest('[data-step]');
        if (!step) return;
        if (event.target.closest('[data-remove-step]')) {
            if (stepsBox.querySelectorAll('[data-step]').length <= 1) return;
            step.remove(); renumber();
        }
        if (event.target.closest('[data-move-up]')) {
            const prev = step.previousElementSibling; if (prev) stepsBox.insertBefore(step, prev); renumber();
        }
        if (event.target.closest('[data-move-down]')) {
            const next = step.nextElementSibling; if (next) stepsBox.insertBefore(next, step); renumber();
        }
    });

    stepsBox.addEventListener('change', event => {
        if (!event.target.matches('[data-approver-type]')) return;
        const step = event.target.closest('[data-step]');
        const select = step.querySelector('[data-approver-select]');
        select.innerHTML = optionsFor(event.target.value);
    });

    renumber();
})();
</script>
