<?php
declare(strict_types=1);

$config = require __DIR__ . '/../config/config.php';

date_default_timezone_set($config['app']['timezone']);

ini_set('session.use_strict_mode', '1');
ini_set('session.use_only_cookies', '1');
ini_set('session.use_trans_sid', '0');
ini_set('session.cookie_httponly', '1');
ini_set('session.cookie_secure', $config['security']['session_secure'] ? '1' : '0');
ini_set('session.cookie_samesite', $config['security']['session_samesite']);
ini_set('session.cookie_lifetime', '0');

header('X-Content-Type-Options: nosniff');
header('X-Frame-Options: SAMEORIGIN');
header('Referrer-Policy: strict-origin-when-cross-origin');

session_name($config['app']['session_name']);
session_start();

require_once __DIR__ . '/helpers.php';
require_once __DIR__ . '/Csrf.php';
require_once __DIR__ . '/ApiException.php';
require_once __DIR__ . '/ApiClient.php';
require_once __DIR__ . '/Auth.php';
require_once __DIR__ . '/View.php';
require_once __DIR__ . '/ListQuery.php';
require_once __DIR__ . '/Paginator.php';

AppContext::init($config);
