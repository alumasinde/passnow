<?php $title = 'Create organization'; ?>
<section class="page-header">
    <div>
        <span class="eyebrow">Platform</span>
        <h1>Create organization</h1>
        <p>Set up a new tenant and its first administrator in one step.</p>
    </div>
    <a class="btn btn-secondary" href="<?= e(url('platform/tenants')) ?>">Back to tenants</a>
</section>

<?php if (!empty($error)): ?>
    <div class="alert alert-danger"><?= e($error) ?></div>
<?php endif ?>

<section class="content-card">
    <form method="post" class="form-stack" autocomplete="off">
        <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">

        <div class="form-section">
            <h2>Organization</h2>
            <p class="muted">This creates the tenant that owns its own visitors, visits, gatepasses, users, and settings.</p>

            <div class="form-grid">
                <div class="field">
                    <label for="tenant_name">Organization name</label>
                    <input id="tenant_name" name="tenant_name" value="<?= e($form['tenant_name']) ?>" required maxlength="160" autofocus>
                </div>

                <div class="field">
                    <label for="tenant_slug">Organization slug</label>
                    <input id="tenant_slug" name="tenant_slug" value="<?= e($form['tenant_slug']) ?>" required minlength="3" maxlength="50" pattern="[a-z0-9-]+">
                    <small>Lowercase letters, numbers, and hyphens only.</small>
                </div>
            </div>
        </div>

        <div class="form-section">
            <h2>First administrator</h2>
            <p class="muted">This user receives the Tenant Admin role with access to configure the organization.</p>

            <div class="form-grid">
                <div class="field">
                    <label for="admin_first_name">First name</label>
                    <input id="admin_first_name" name="admin_first_name" value="<?= e($form['admin_first_name']) ?>" required maxlength="100">
                </div>

                <div class="field">
                    <label for="admin_last_name">Last name</label>
                    <input id="admin_last_name" name="admin_last_name" value="<?= e($form['admin_last_name']) ?>" required maxlength="100">
                </div>

                <div class="field field-span-2">
                    <label for="admin_email">Email address</label>
                    <input id="admin_email" name="admin_email" type="email" value="<?= e($form['admin_email']) ?>" required maxlength="255">
                </div>

                <div class="field">
                    <label for="admin_password">Password</label>
                    <input id="admin_password" name="admin_password" type="password" required minlength="12">
                    <small>Use at least 12 characters.</small>
                </div>

                <div class="field">
                    <label for="admin_password_confirmation">Confirm password</label>
                    <input id="admin_password_confirmation" name="admin_password_confirmation" type="password" required minlength="12">
                </div>
            </div>
        </div>

        <div class="form-actions">
            <a class="btn btn-secondary" href="<?= e(url('platform/tenants')) ?>">Cancel</a>
            <button class="btn btn-primary" type="submit">Create organization</button>
        </div>
    </form>
</section>

