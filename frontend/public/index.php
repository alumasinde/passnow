<?php declare(strict_types=1); require_once __DIR__.'/../app/App.php'; Auth::requireLogin(); redirect('dashboard.php');
