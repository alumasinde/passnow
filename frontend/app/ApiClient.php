<?php

declare(strict_types=1);
final class ApiClient
{
    public function __construct(private readonly string $baseUrl, private readonly int $timeout = 30) {}
    public function request(string $method, string $path, ?array $body = null, ?string $accessToken = null): array
    {
        $requestPath = '/' . ltrim($path, '/');
        $tenantHost = (string)AppContext::config('app.tenant_host', '');
        $baseDomain = strtolower((string)(getenv('BASE_DOMAIN') ?: ''));
        $currentPath = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
        $segments = array_values(array_filter(explode('/', trim($currentPath, '/'))));
        $reserved = ['login', 'logout', 'dashboard', 'platform', 'assets', 'api'];
        $localTenantSlug = strtolower(trim((string)($_SESSION['tenant_slug'] ?? '')));
        if ($localTenantSlug === '' && $segments !== [] && !in_array(strtolower($segments[0]), $reserved, true) && !str_ends_with(strtolower($segments[0]), '.php')) {
            $localTenantSlug = strtolower($segments[0]);
        }

        // Local development can use /{tenant}/login in the browser, but the
        // Go API routes themselves remain /api/v1/.... Resolve the tenant with
        // X-Tenant-Slug instead of prefixing the API path. Prefixing here would
        // make the router receive /{tenant}/api/v1/... and return an HTML 404.
        if ($localTenantSlug !== '' && ($baseDomain === '' || stripos($tenantHost, $baseDomain) === false)) {
            $tenantHost = '';
        }

        $url = $this->baseUrl . $requestPath;
        $headers = ['Accept: application/json', 'Content-Type: application/json'];
        if ($tenantHost !== '') $headers[] = 'Host: ' . $tenantHost;
        if ($localTenantSlug !== '') $headers[] = 'X-Tenant-Slug: ' . $localTenantSlug;
        if ($accessToken) $headers[] = 'Authorization: Bearer ' . $accessToken;
        $ch = curl_init($url);
        curl_setopt_array($ch, [CURLOPT_CUSTOMREQUEST => strtoupper($method), CURLOPT_RETURNTRANSFER => true, CURLOPT_FOLLOWLOCATION => false, CURLOPT_CONNECTTIMEOUT => min($this->timeout, 5), CURLOPT_TIMEOUT => $this->timeout, CURLOPT_HTTPHEADER => $headers, CURLOPT_SSL_VERIFYPEER => true, CURLOPT_SSL_VERIFYHOST => 2]);
        if ($body !== null) curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($body, JSON_THROW_ON_ERROR));
        $raw = curl_exec($ch);
        if ($raw === false) {
            $err = curl_error($ch);
            $errno = curl_errno($ch);
            curl_close($ch);

            throw new ApiException(
                'The API could not be reached: ' . ($err !== '' ? $err : 'Unknown cURL transport error.'),
                0,
                ['transport_error' => $err, 'transport_errno' => $errno, 'url' => $url]
            );
        }
        $status = (int)curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);
        $decoded = json_decode($raw, true);
        if (!is_array($decoded)) {
            throw new ApiException('The API returned an invalid response (HTTP ' . $status . ').', $status);
        }
        if ($status < 200 || $status >= 300) {
            throw new ApiException(self::errorMessage($decoded), $status, $decoded + ['request_url' => $url]);
        }
        return $decoded;
    }

    public function requestMultipart(string $method, string $path, array $fields, array $files, ?string $accessToken = null): array
    {
        $requestPath = '/' . ltrim($path, '/');
        $tenantHost = (string)AppContext::config('app.tenant_host', '');
        $baseDomain = strtolower((string)(getenv('BASE_DOMAIN') ?: ''));
        $currentPath = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
        $segments = array_values(array_filter(explode('/', trim($currentPath, '/'))));
        $reserved = ['login', 'logout', 'dashboard', 'platform', 'assets', 'api'];
        $localTenantSlug = strtolower(trim((string)($_SESSION['tenant_slug'] ?? '')));
        if ($localTenantSlug === '' && $segments !== [] && !in_array(strtolower($segments[0]), $reserved, true) && !str_ends_with(strtolower($segments[0]), '.php')) {
            $localTenantSlug = strtolower($segments[0]);
        }
        if ($localTenantSlug !== '' && ($baseDomain === '' || stripos($tenantHost, $baseDomain) === false)) $tenantHost = '';

        $url = $this->baseUrl . $requestPath;
        $headers = ['Accept: application/json'];
        if ($tenantHost !== '') $headers[] = 'Host: ' . $tenantHost;
        if ($localTenantSlug !== '') $headers[] = 'X-Tenant-Slug: ' . $localTenantSlug;
        if ($accessToken) $headers[] = 'Authorization: Bearer ' . $accessToken;

        $payload = $fields;
        foreach ($files as $field => $pathValue) {
            if (!is_string($pathValue) || $pathValue === '' || !is_file($pathValue)) continue;
            $payload[$field] = new CURLFile($pathValue);
        }

        $ch = curl_init($url);
        curl_setopt_array($ch, [
            CURLOPT_CUSTOMREQUEST => strtoupper($method),
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => false,
            CURLOPT_CONNECTTIMEOUT => min($this->timeout, 5),
            CURLOPT_TIMEOUT => $this->timeout,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_SSL_VERIFYPEER => true,
            CURLOPT_SSL_VERIFYHOST => 2,
            CURLOPT_POSTFIELDS => $payload,
        ]);
        $raw = curl_exec($ch);
        if ($raw === false) {
            $err = curl_error($ch); $errno = curl_errno($ch); curl_close($ch);
            throw new ApiException('The API could not be reached: ' . ($err !== '' ? $err : 'Unknown cURL transport error.'), 0, ['transport_error'=>$err,'transport_errno'=>$errno,'url'=>$url]);
        }
        $status = (int)curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);
        $decoded = json_decode($raw, true);
        if (!is_array($decoded)) throw new ApiException('The API returned an invalid response (HTTP ' . $status . ').', $status);
        if ($status < 200 || $status >= 300) {
            throw new ApiException(self::errorMessage($decoded), $status, $decoded + ['request_url'=>$url]);
        }
        return $decoded;
    }

    public function requestBinary(string $method, string $path, ?string $accessToken = null): array
    {
        $requestPath = '/' . ltrim($path, '/');
        $tenantHost = (string)AppContext::config('app.tenant_host', '');
        $currentPath = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH) ?: '/';
        $segments = array_values(array_filter(explode('/', trim($currentPath, '/'))));
        $reserved = ['login','logout','dashboard','platform','assets','api'];
        $slug = strtolower(trim((string)($_SESSION['tenant_slug'] ?? '')));
        if ($slug === '' && $segments !== [] && !in_array(strtolower($segments[0]), $reserved, true)) $slug = strtolower($segments[0]);
        $url = $this->baseUrl . $requestPath;
        $headers = ['Accept: image/png'];
        if ($tenantHost !== '') $headers[] = 'Host: ' . $tenantHost;
        if ($slug !== '') $headers[] = 'X-Tenant-Slug: ' . $slug;
        if ($accessToken) $headers[] = 'Authorization: Bearer ' . $accessToken;
        $ch=curl_init($url);
        curl_setopt_array($ch,[CURLOPT_CUSTOMREQUEST=>strtoupper($method),CURLOPT_RETURNTRANSFER=>true,CURLOPT_FOLLOWLOCATION=>false,CURLOPT_CONNECTTIMEOUT=>min($this->timeout,5),CURLOPT_TIMEOUT=>$this->timeout,CURLOPT_HTTPHEADER=>$headers,CURLOPT_SSL_VERIFYPEER=>true,CURLOPT_SSL_VERIFYHOST=>2]);
        $raw=curl_exec($ch);
        if($raw===false){$err=curl_error($ch);curl_close($ch);throw new ApiException('The API could not be reached: '.($err?:'Unknown transport error.'));}
        $status=(int)curl_getinfo($ch,CURLINFO_RESPONSE_CODE);$type=(string)curl_getinfo($ch,CURLINFO_CONTENT_TYPE);curl_close($ch);
        if($status<200||$status>=300)throw new ApiException('Unable to load binary API response (HTTP '.$status.').',$status);
        return ['body'=>$raw,'content_type'=>$type?:'application/octet-stream'];
    }

    private static function errorMessage(array $payload): string
    {
        $candidate = $payload['message'] ?? $payload['error'] ?? null;

        if (is_array($candidate)) {
            foreach (['message', 'detail', 'error', 'code'] as $key) {
                if (isset($candidate[$key]) && is_scalar($candidate[$key])) {
                    return trim((string)$candidate[$key]) ?: 'The request could not be completed.';
                }
            }
            foreach ($candidate as $value) {
                if (is_scalar($value) && trim((string)$value) !== '') {
                    return trim((string)$value);
                }
            }
        }

        if (is_scalar($candidate) && trim((string)$candidate) !== '') {
            return trim((string)$candidate);
        }

        return 'The request could not be completed.';
    }
}
