<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();
$q=new ListQuery(['status','date']);
$extra=[];foreach(['status','date'] as $k)if(isset($_GET[$k])&&$_GET[$k]!=='')$extra[$k]=trim((string)$_GET[$k]);
$r=ResourcePage::list('/api/v1/visits',$q,$extra);
$statusOptions=is_array($r['meta']['statuses']??null)?$r['meta']['statuses']:[];
$columns=[['key'=>'badge_number','label'=>'Badge'],['key'=>'visitor_name','label'=>'Visitor'],['key'=>'host_name','label'=>'Host'],['key'=>'purpose','label'=>'Purpose'],['key'=>'status','label'=>'Status'],['key'=>'expected_time','label'=>'Expected']];
$actions=[['label'=>'View','icon'=>'fa-eye','class'=>'btn-secondary','href'=>fn($row)=>url('visit.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('visits/index',['query'=>$q,'rows'=>$r['rows'],'paginator'=>$r['paginator'],'columns'=>$columns,'rowActions'=>$actions,'error'=>$r['error'],'statusOptions'=>$statusOptions]);
