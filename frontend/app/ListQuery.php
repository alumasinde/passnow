<?php
declare(strict_types=1);

final class ListQuery
{
    private array $allowedKeys;

    public function __construct(array $allowedKeys = [])
    {
        $this->allowedKeys = $allowedKeys;
    }

    public function fromRequest(): array
    {
        $query = [];

        foreach ($this->allowedKeys as $key) {
            if (!isset($_GET[$key])) continue;

            $value = is_string($_GET[$key]) ? trim($_GET[$key]) : '';
            if ($value !== '') $query[$key] = $value;
        }

        return $query;
    }

    public function page(int $default = 1): int
    {
        $value = filter_input(INPUT_GET, 'page', FILTER_VALIDATE_INT);
        return max(1, $value ?: $default);
    }

    public function perPage(int $default = 20): int
    {
        $value = filter_input(INPUT_GET, 'per_page', FILTER_VALIDATE_INT);
        $allowed = [10, 20, 50, 100];

        return in_array($value, $allowed, true) ? $value : $default;
    }

    public function search(): string
    {
        return trim((string) ($_GET['q'] ?? ''));
    }

    public function url(array $changes = []): string
    {
        $params = $_GET;

        foreach ($changes as $key => $value) {
            if ($value === null || $value === '') {
                unset($params[$key]);
            } else {
                $params[$key] = $value;
            }
        }

        $query = http_build_query($params);
        return basename($_SERVER['PHP_SELF'] ?? '') . ($query ? '?' . $query : '');
    }
}
