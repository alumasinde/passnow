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
        $localTenantSlug = '';
        if ($segments !== [] && !in_array(strtolower($segments[0]), $reserved, true) && !str_ends_with(strtolower($segments[0]), '.php')) {
            $localTenantSlug = strtolower($segments[0]);
        }

        // Local development can use /{tenant}/login while production can use
        // a tenant subdomain/custom domain. Prefix API requests only for the
        // path-based local mode.
        if ($localTenantSlug !== '' && ($baseDomain === '' || stripos($tenantHost, $baseDomain) === false)) {
            $requestPath = '/' . $localTenantSlug . $requestPath;
            $tenantHost = '';
        }

        $url = $this->baseUrl . $requestPath;
        $headers = ['Accept: application/json', 'Content-Type: application/json'];
        if ($tenantHost !== '') $headers[] = 'Host: ' . $tenantHost;
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
            $messageValue = $decoded['message'] ?? $decoded['error'] ?? 'The request could not be completed.';

            if (is_array($messageValue)) {
                $parts = [];
                foreach ($messageValue as $key => $value) {
                    if (is_array($value)) {
                        $value = implode(', ', array_map(
                            static fn ($item): string => is_scalar($item) || $item === null ? (string) $item : json_encode($item),
                            $value
                        ));
                    }
                    $parts[] = is_string($key) ? $key . ': ' . (string) $value : (string) $value;
                }
                $message = $parts !== [] ? implode(' | ', $parts) : 'The request could not be completed.';
            } elseif (is_scalar($messageValue) || $messageValue === null) {
                $message = (string) $messageValue;
            } else {
                $message = 'The request could not be completed.';
            }

            throw new ApiException($message, $status, $decoded + ['request_url' => $url]);
        }
        return $decoded;
    }
}
