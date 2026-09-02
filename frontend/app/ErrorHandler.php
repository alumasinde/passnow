<?php
declare(strict_types=1);

final class ErrorHandler
{
    private static bool $rendering = false;
    private static bool $debug = false;

    public static function register(array $config): void
    {
        self::$debug = (bool)($config['errors']['debug'] ?? false);

        ini_set('display_errors', '0');
        ini_set('display_startup_errors', '0');
        ini_set('log_errors', '1');

        if (ob_get_level() === 0) {
            ob_start();
        }

        set_exception_handler(static function (Throwable $error): void {
            self::renderThrowable($error);
        });

        register_shutdown_function(static function (): void {
            $error = error_get_last();
            if (!is_array($error)) return;

            $fatalTypes = [E_ERROR, E_PARSE, E_CORE_ERROR, E_COMPILE_ERROR, E_USER_ERROR, E_RECOVERABLE_ERROR];
            if (!in_array((int)($error['type'] ?? 0), $fatalTypes, true)) return;

            $message = (string)($error['message'] ?? 'Fatal application error');
            $file = (string)($error['file'] ?? '');
            $line = (int)($error['line'] ?? 0);
            self::renderFatal($message, $file, $line);
        });
    }

    private static function renderThrowable(Throwable $error): void
    {
        error_log(sprintf(
            '[PassNow][frontend][uncaught] %s in %s:%d — %s',
            get_class($error),
            $error->getFile(),
            $error->getLine(),
            $error->getMessage()
        ));

        self::render(
            500,
            self::$debug ? $error->getMessage() : null,
            self::$debug ? self::formatThrowable($error) : null
        );
    }

    private static function renderFatal(string $message, string $file, int $line): void
    {
        error_log(sprintf('[PassNow][frontend][fatal] %s in %s:%d', $message, $file, $line));

        self::render(
            500,
            self::$debug ? $message : null,
            self::$debug ? ($file . ':' . $line) : null
        );
    }

    private static function render(int $status, ?string $debugMessage, ?string $debugDetails): void
    {
        if (self::$rendering) {
            self::fallback($status);
            return;
        }

        self::$rendering = true;

        while (ob_get_level() > 0) {
            ob_end_clean();
        }

        if (!headers_sent()) {
            http_response_code($status);
        }

        $errorId = strtoupper(bin2hex(random_bytes(4)));

        try {
            View::render('errors/500', [
                'status' => $status,
                'errorId' => $errorId,
                'debugMessage' => $debugMessage,
                'debugDetails' => $debugDetails,
                'title' => 'Application error',
            ], 'error');
        } catch (Throwable $renderError) {
            error_log('[PassNow][frontend][error-render-fallback] ' . $renderError->getMessage());
            self::fallback($status);
        }

        exit;
    }

    private static function formatThrowable(Throwable $error): string
    {
        return get_class($error) . ': ' . $error->getMessage()
            . "\n" . $error->getFile() . ':' . $error->getLine()
            . "\n\n" . $error->getTraceAsString();
    }

    private static function fallback(int $status): void
    {
        http_response_code($status);
        echo '<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>PassNow</title></head><body style="font-family:Arial,sans-serif;background:#f5f7fb;color:#182230;display:grid;place-items:center;min-height:100vh;margin:0"><main style="max-width:520px;padding:32px;background:#fff;border:1px solid #e4e7ec;border-radius:16px;text-align:center"><div style="font-size:48px;font-weight:800;color:#2563eb">500</div><h1>Something went wrong</h1><p>Please refresh the page or return to the dashboard.</p></main></body></html>';
        exit;
    }
}
