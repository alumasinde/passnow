<!doctype html>
<html lang="en">
<head>
<?php $platformLayout = true; require __DIR__ . '/../partials/head.php'; ?>
</head>
<body class="app-page platform-page">
<div class="app-shell">
    <?php require __DIR__ . '/../partials/platform-sidebar.php'; ?>
    <div class="app-main">
        <?php require __DIR__ . '/../partials/platform-topbar.php'; ?>
        <main class="page-content">
            <?php require __DIR__ . '/../partials/flash.php'; ?>
            <?php require $viewFile; ?>
        </main>
        <?php require __DIR__ . '/../partials/footer.php'; ?>
    </div>
</div>
<?php require __DIR__ . '/../partials/scripts.php'; ?>
</body>
</html>
