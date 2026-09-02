<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery(['status','department_id']);$extra=[];
foreach(['status','department_id'] as $k)if(isset($_GET[$k])&&$_GET[$k]!=='')$extra[$k]=trim((string)$_GET[$k]);
$r=ResourcePage::list('/api/v1/employees',$q,$extra);
$columns=[['key'=>'employee_number','label'=>'Employee no.'],['key'=>'full_name','label'=>'Employee'],['key'=>'department_name','label'=>'Department'],['key'=>'phone','label'=>'Phone'],['key'=>'status','label'=>'Status']];
$actions=[['label'=>'View','icon'=>'fa-eye','class'=>'btn-secondary','href'=>fn($row)=>url('employee.php?id='.rawurlencode((string)($row['id']??''))) ]];
$statusOptions=is_array($r['meta']['statuses']??null)?$r['meta']['statuses']:[];
$departmentOptions=is_array($r['meta']['departments']??null)?$r['meta']['departments']:[];
App::render('employees/index',['query'=>$q,'rows'=>$r['rows'],'paginator'=>$r['paginator'],'columns'=>$columns,'rowActions'=>$actions,'error'=>$r['error'],'statusOptions'=>$statusOptions,'departmentOptions'=>$departmentOptions]);
