<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$q = new ListQuery(['status','blacklisted']);
$extra=[];
foreach(['status','blacklisted'] as $key){if(isset($_GET[$key])&&$_GET[$key]!=='')$extra[$key]=trim((string)$_GET[$key]);}
$r=ResourcePage::list('/api/v1/visitors',$q,$extra);
$meta=$r['meta'];
$columns=[
 ['key'=>'full_name','label'=>'Visitor'],
 ['key'=>'phone','label'=>'Phone'],
 ['key'=>'id_number','label'=>'ID number'],
 ['key'=>'company_name','label'=>'Company'],
 ['key'=>'status','label'=>'Status'],
];
$actions=[['label'=>'View','icon'=>'fa-eye','class'=>'btn-secondary','href'=>fn($row)=>url('visitor.php?id='.rawurlencode((string)($row['id']??''))) ]];
$statusOptions=is_array($meta['statuses']??null)?$meta['statuses']:[];
$blacklistOptions=[['value'=>'0','label'=>'Not blacklisted'],['value'=>'1','label'=>'Blacklisted']];
App::render('visitors/index',compact('q','r','meta','columns','actions','statusOptions','blacklistOptions')+[
 'query'=>$q,'rows'=>$r['rows'],'paginator'=>$r['paginator'],'rowActions'=>$actions,'error'=>$r['error']
]);
