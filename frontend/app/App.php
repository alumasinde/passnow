<?php
declare(strict_types=1);
require_once __DIR__.'/bootstrap.php';
final class App { private static ?ApiClient $api=null; public static function api():ApiClient{return self::$api??=new ApiClient((string)AppContext::config('app.api_base_url'),(int)AppContext::config('security.api_timeout',15));} public static function render(string $view,array $data=[],string $layout='app'):void{View::render($view,$data,$layout);} public static function config(?string $key=null,mixed $default=null):mixed{return AppContext::config($key,$default);} }

/** Store a one-time UI flash message in the current session. */
function flash(string $type, string $message): void {
    if (session_status() !== PHP_SESSION_ACTIVE) session_start();
    $_SESSION['_flash'] = ['type'=>$type, 'message'=>$message];
}
/** Consume the current one-time UI flash message. */
function flash_get(): ?array {
    if (session_status() !== PHP_SESSION_ACTIVE) session_start();
    $value = $_SESSION['_flash'] ?? null;
    unset($_SESSION['_flash']);
    return is_array($value) ? $value : null;
}
