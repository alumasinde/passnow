<?php
declare(strict_types=1);
final class ConfigCrud
{
    public static function create(string $endpoint, array $payload): array
    {
        return Auth::api(App::api(),'POST',$endpoint,$payload);
    }
    public static function update(string $endpoint, array $payload): array
    {
        return Auth::api(App::api(),'PATCH',$endpoint,$payload);
    }
}
