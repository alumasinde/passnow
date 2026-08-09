<?php
declare(strict_types=1);

final class AdminResource
{
    public static function page(string $endpoint, ListQuery $query, array $extra=[]): array
    {
        return ResourcePage::list($endpoint,$query,$extra);
    }

    public static function options(array $rows): array
    {
        return array_values(array_map(static function(array $row): array {
            return [
                'value'=>(string)($row['id'] ?? $row['value'] ?? ''),
                'label'=>(string)($row['name'] ?? $row['label'] ?? $row['title'] ?? $row['description'] ?? ''),
            ];
        }, $rows));
    }
}
