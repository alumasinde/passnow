<main class="error-screen">
    <section class="error-card" role="alert" aria-labelledby="error-title">
        <div class="error-icon" aria-hidden="true"><i class="fa-solid fa-triangle-exclamation"></i></div>
        <div class="error-status"><?= (int)($status ?? 500) ?></div>
        <h1 id="error-title">Something went wrong</h1>
        <p>PassNow could not complete this request. Please try again. Your data has not been changed by this error.</p>

        <?php if (!empty($errorId)): ?>
            <div class="error-reference"><span>Error reference</span><code><?= e((string)$errorId) ?></code></div>
        <?php endif; ?>

        <div class="error-actions">
            <button class="btn btn-primary" type="button" onclick="window.location.reload()">Try again</button>
            <a class="btn btn-secondary" href="<?= e(url('dashboard')) ?>">Go to dashboard</a>
        </div>

        <?php if (!empty($debugMessage)): ?>
            <details class="error-debug">
                <summary>Technical details</summary>
                <strong><?= e((string)$debugMessage) ?></strong>
                <?php if (!empty($debugDetails)): ?><pre><?= e((string)$debugDetails) ?></pre><?php endif; ?>
            </details>
        <?php endif; ?>
    </section>
</main>
