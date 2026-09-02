<?php $theme = Theme::current(); ?>
<section class="auth-card">
    <div class="auth-brand">
        <span class="brand-mark">
            <?php if (($theme['logo_url'] ?? '') !== ''): ?>
                <img src="<?= e((string)$theme['logo_url']) ?>" alt="<?= e(Theme::brandName()) ?> logo" class="brand-logo">
            <?php else: ?>
                <i class="fa-solid fa-door-open"></i>
            <?php endif; ?>
        </span>
        <div><strong><?= e(Theme::brandName()) ?></strong><span>Gatepass management</span></div>
    </div>
    <div class="auth-heading"><h1>Welcome back</h1><p>Sign in to continue to your workspace.</p></div>
    <?php if (!empty($errors)): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><div><?php foreach ($errors as $error): ?><div><?= e($error) ?></div><?php endforeach; ?></div></div><?php endif; ?>
    <form method="post" class="form-stack" data-loading-form>
        <input type="hidden" name="_csrf" value="<?= e(Csrf::token()) ?>">
        <div class="field"><label for="email">Email address</label><div class="input-icon"><i class="fa-regular fa-envelope"></i><input id="email" name="email" type="email" autocomplete="username" value="<?= e($email ?? '') ?>" required maxlength="190"></div></div>
        <div class="field"><label for="password">Password</label><div class="input-icon"><i class="fa-solid fa-lock"></i><input id="password" name="password" type="password" autocomplete="current-password" required maxlength="200"><button class="input-action" type="button" data-toggle-password="#password" aria-label="Show password"><i class="fa-regular fa-eye"></i></button></div></div>
        <button class="btn btn-primary btn-block" type="submit" data-loading-label="Signing in..."><span data-button-label>Sign in</span></button>
    </form>
</section>
