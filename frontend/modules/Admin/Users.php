<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery(['status','role_id']);$extra=[];
foreach(['status','role_id'] as $k)if(isset($_GET[$k])&&$_GET[$k]!=='')$extra[$k]=trim((string)$_GET[$k]);
$r=AdminResource::page('/api/v1/users',$q,$extra);
$columns=[['key'=>'email','label'=>'Email'],['key'=>'full_name','label'=>'Name'],['key'=>'role_name','label'=>'Role'],['key'=>'status','label'=>'Status']];
$actions=[['label'=>'View','icon'=>'fa-eye','class'=>'btn-secondary','href'=>fn($row)=>url('user.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/users',compact('q','r','columns','actions'));
