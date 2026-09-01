<?php

declare(strict_types=1);
final class ApiClient
{
    public function __construct(private readonly string $baseUrl, private readonly int $timeout = 15) {}
    public function request(string $method, string $path, ?array $body = null, ?string $accessToken = null): array
    {
        $url = $this->baseUrl . '/' . ltrim($path, '/');
        $headers = ['Accept: application/json', 'Content-Type: application/json'];
        $tenantHost = (string)AppContext::config('app.tenant_host', '');
        if ($tenantHost !== '') $headers[] = 'Host: ' . $tenantHost;
        if ($accessToken) $headers[] = 'Authorization: Bearer ' . $accessToken;
        $ch = curl_init($url);
        curl_setopt_array($ch, [CURLOPT_CUSTOMREQUEST => strtoupper($method), CURLOPT_RETURNTRANSFER => true, CURLOPT_FOLLOWLOCATION => false, CURLOPT_CONNECTTIMEOUT => min($this->timeout, 5), CURLOPT_TIMEOUT => $this->timeout, CURLOPT_HTTPHEADER => $headers, CURLOPT_SSL_VERIFYPEER => true, CURLOPT_SSL_VERIFYHOST => 2]);
        if ($body !== null) curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($body, JSON_THROW_ON_ERROR));
        $raw = curl_exec($ch);
        if ($raw === false) {
            $err = curl_error($ch);
            curl_close($ch);
            throw new ApiException('The API could not be reached.', 0, ['transport_error' => $err]);
        }
        $status = (int)curl_getinfo($ch, CURLINFO_RESPONSE_CODE);
        curl_close($ch);
        $decoded = json_decode($raw, true);
        if (!is_array($decoded)) throw new ApiException('The API returned an invalid response.', $status);
        if ($status < 200 || $status >= 300) throw new ApiException((string)($decoded['message'] ?? 'The request could not be completed.'), $status, $decoded);
        return $decoded;
    }
}
