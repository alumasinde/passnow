<?php
declare(strict_types=1);

final class ApiClient
{
    public function __construct(
        private readonly string $baseUrl,
        private readonly int $timeout = 15
    ) {
        if (!filter_var($this->baseUrl, FILTER_VALIDATE_URL) || !preg_match('~^https?://~i', $this->baseUrl)) {
            throw new InvalidArgumentException('API_BASE_URL must be a valid HTTP(S) URL.');
        }
    }

    public function request(
        string $method,
        string $path,
        ?array $body = null,
        ?string $accessToken = null
    ): array {
        $method = strtoupper(trim($method));
        if (!preg_match('/^[A-Z]+$/', $method)) {
            throw new InvalidArgumentException('Invalid HTTP method.');
        }

        if ($path === '' || preg_match('~^(?:[a-z][a-z0-9+.-]*:|//)~i', $path)) {
            throw new InvalidArgumentException('API paths must be relative paths.');
        }

        $url = $this->baseUrl . '/' . ltrim($path, '/');
        $headers = [
            'Accept: application/json',
            'Content-Type: application/json',
            'User-Agent: PassNow-Frontend/1.0',
            'X-Request-ID: ' . bin2hex(random_bytes(16)),
        ];

        $tenantHost = $this->validatedTenantHost((string) AppContext::config('app.tenant_host', ''));
        if ($tenantHost !== '') {
            $headers[] = 'Host: ' . $tenantHost;
        }

        if ($accessToken !== null && $accessToken !== '') {
            $headers[] = 'Authorization: Bearer ' . $accessToken;
        }

        $ch = curl_init($url);
        if ($ch === false) {
            throw new ApiException('The API could not be reached.', 0);
        }

        $encodedBody = null;
        if ($body !== null) {
            try {
                $encodedBody = json_encode($body, JSON_THROW_ON_ERROR | JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            } catch (JsonException) {
                curl_close($ch);
                throw new ApiException('The request could not be encoded.', 0);
            }
        }

        curl_setopt_array($ch, [
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => false,
            CURLOPT_CONNECTTIMEOUT => min($this->timeout, 5),
            CURLOPT_TIMEOUT => $this->timeout,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_SSL_VERIFYPEER => true,
            CURLOPT_SSL_VERIFYHOST => 2,
        ]);

        if ($encodedBody !== null) {
            curl_setopt($ch, CURLOPT_POSTFIELDS, $encodedBody);
        }

        $raw = curl_exec($ch);
        if ($raw === false) {
            curl_close($ch);
            throw new ApiException('The API could not be reached.', 0);
        }

        $status = (int) curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);

        if ($status === 204 || trim($raw) === '') {
            if ($status >= 200 && $status < 300) return [];
            throw new ApiException('The request could not be completed.', $status);
        }

        try {
            $decoded = json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
        } catch (JsonException) {
            throw new ApiException(
                $status >= 500
                    ? 'The API is temporarily unavailable.'
                    : 'The API returned an invalid response.',
                $status
            );
        }

        if (!is_array($decoded)) {
            throw new ApiException('The API returned an invalid response.', $status);
        }

        if ($status < 200 || $status >= 300) {
            $message = (string) ($decoded['message'] ?? 'The request could not be completed.');
            throw new ApiException($message, $status, $decoded);
        }

        return $decoded;
    }

    private function validatedTenantHost(string $host): string
    {
        $host = trim($host);
        if ($host === '') return '';
        if (strlen($host) > 255 || preg_match('/[\r\n]/', $host)) {
            throw new ApiException('The tenant host is invalid.', 400);
        }

        $parsed = parse_url('http://' . $host);
        $hostname = strtolower((string) ($parsed['host'] ?? ''));
        $port = isset($parsed['port']) ? (int) $parsed['port'] : null;

        if ($hostname === '' || !filter_var($hostname, FILTER_VALIDATE_DOMAIN, FILTER_FLAG_HOSTNAME)) {
            if ($hostname !== 'localhost' && filter_var($hostname, FILTER_VALIDATE_IP) === false) {
                throw new ApiException('The tenant host is invalid.', 400);
            }
        }

        if ($port !== null && ($port < 1 || $port > 65535)) {
            throw new ApiException('The tenant host is invalid.', 400);
        }

        return $hostname . ($port !== null ? ':' . $port : '');
    }
}
