<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';
Auth::requireLogin();

$query = new ListQuery(['status']);
$page = $query->page();
$perPage = $query->perPage((int)App::config('ui.page_size', 20));

$params = ['page'=>$page,'per_page'=>$perPage];
if ($query->search() !== '') $params['q'] = $query->search();
if (!empty($_GET['status'])) $params['status'] = trim((string)$_GET['status']);

$rows=[]; $total=0; $error=null; $meta=[];
try {
    $payload=Auth::api(App::api(),'GET','/api/v1/approvals/pending?'.http_build_query($params));
    $rows=apiRows($payload); $meta=apiMeta($payload);
    $total=(int)($meta['total'] ?? $payload['total'] ?? count($rows));
} catch (ApiException $e) { $error=$e->getMessage(); }
  catch (Throwable) { $error='Pending approvals are temporarily unavailable.'; }

$paginator=new Paginator($page,$perPage,$total);
$statusOptions=is_array($meta['statuses'] ?? null) ? $meta['statuses'] : [];

$columns=[
 ['key'=>'gatepass_number','label'=>'Gatepass'],
 ['key'=>'gatepass_type_name','label'=>'Type'],
 ['key'=>'subject_name','label'=>'Person'],
 ['key'=>'step_name','label'=>'Approval step'],
 ['key'=>'status','label'=>'Status'],
 ['key'=>'created_at','label'=>'Submitted'],
];
$rowActions=[[
 'label'=>'Review','icon'=>'fa-eye','class'=>'btn-primary',
 'href'=>static fn(array $row):string=>url('approval.php?id='.rawurlencode((string)($row['gatepass_id'] ?? $row['id'] ?? '')))
]];

App::render('approvals/index',compact('query','rows','paginator','columns','rowActions','error','statusOptions'));
