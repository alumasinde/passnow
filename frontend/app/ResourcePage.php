<?php
declare(strict_types=1);

final class ResourcePage
{
    public static function list(
        string $endpoint,
        ListQuery $query,
        array $extraParams = [],
        ?callable $normalizer = null
    ): array {
        $page = $query->page();
        $perPage = $query->perPage((int)App::config('ui.page_size', 20));
        $params = array_merge(['limit'=>$perPage,'offset'=>($page-1)*$perPage], $extraParams);
        if ($query->search() !== '') $params['q'] = $query->search();

        try {
            $payload = Auth::api(App::api(), 'GET', $endpoint . '?' . http_build_query($params));
            $rows = apiRows($payload);
            if ($normalizer) $rows = $normalizer($rows);
            $meta = apiMeta($payload);
            $total = (int)($meta['total'] ?? $payload['total'] ?? count($rows));
            return [
                'rows'=>$rows, 'meta'=>$meta, 'total'=>$total, 'error'=>null,
                'paginator'=>new Paginator($page,$perPage,$total),
            ];
        } catch (ApiException $e) {
            return ['rows'=>[],'meta'=>[],'total'=>0,'error'=>$e->getMessage(),
                'paginator'=>new Paginator($page,$perPage,0)];
        } catch (Throwable) {
            return ['rows'=>[],'meta'=>[],'total'=>0,'error'=>'The requested records are temporarily unavailable.',
                'paginator'=>new Paginator($page,$perPage,0)];
        }
    }
}
