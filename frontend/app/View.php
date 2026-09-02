<?php
declare(strict_types=1);

final class View
{
    public static function render(string $view, array $data = [], string $layout = 'app'): void
    {
        $viewFile = __DIR__ . '/../views/' . trim($view, '/') . '.php';
        $layoutFile = __DIR__ . '/../templates/layouts/' . $layout . '.php';

        if (!is_file($viewFile)) {
            throw new RuntimeException('View not found: ' . $view);
        }
        if (!is_file($layoutFile)) {
            throw new RuntimeException('Layout not found: ' . $layout);
        }

        extract($data, EXTR_SKIP);
        require $layoutFile;
    }

    public static function component(string $name, array $data = []): void
    {
        $file = __DIR__ . '/../templates/components/' . trim($name, '/') . '.php';
        if (!is_file($file)) return;
        extract($data, EXTR_SKIP);
        require $file;
    }
}

function component(string $name, array $data = []): void
{
    View::component($name, $data);
}
