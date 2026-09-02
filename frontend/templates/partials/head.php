<?php
$isPlatformLayout = !empty($platformLayout);
if ($isPlatformLayout) {
    $pageTitle = (string)($title ?? 'Platform Administration');
    $brandName = (string)App::config('app.name');
    $themeFavicon = '';
} else {
    $themeFavicon = Theme::faviconURL();
    $brandName = Theme::brandName();
    $pageTitle = (string)($title ?? 'Application');
}
?>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="csrf-token" content="<?= e(Csrf::token()) ?>">
<meta name="color-scheme" content="light dark">
<title><?= e($pageTitle . ' · ' . $brandName) ?></title>
<?php if ($themeFavicon !== ''): ?><link rel="icon" href="<?= e($themeFavicon) ?>"><?php endif; ?>
<link rel="stylesheet" href="<?= e((string)App::config('ui.font_awesome_url')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/variables.css')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/app.css')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/components.css')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/responsive.css')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/auth.css')) ?>">
<link rel="stylesheet" href="<?= e(asset('css/admin.css')) ?>">
<?php if (!$isPlatformLayout): ?><?= Theme::styleTag() ?><?php endif; ?>