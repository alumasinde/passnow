<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery(['status','role_id']);$extra=[];
foreach(['status','role_id'] as $k)if(isset($_GET[$k])&&$_GET[$k]!=='')$extra[$k]=trim((string)$_GET[$k]);
$r=ResourcePage::list('/api/v1/users',$q,$extra,static function(array $rows):array{
 foreach($rows as &$row){$row['id']=$row['user_id']??$row['id']??null;$row['full_name']=trim((string)($row['first_name']??'').' '.(string)($row['last_name']??''));}
 unset($row);return $rows;
});
$columns=[['key'=>'email','label'=>'Email'],['key'=>'full_name','label'=>'Name'],['key'=>'role_name','label'=>'Role'],['key'=>'department_name','label'=>'Department'],['key'=>'status','label'=>'Status']];
$actions=[['label'=>'View','icon'=>'fa-eye','class'=>'btn-secondary','href'=>fn($row)=>url('admin/users/view?user_id='.rawurlencode((string)($row['user_id']??$row['id']??'')))],['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('admin/users/edit?user_id='.rawurlencode((string)($row['user_id']??$row['id']??''))) ]];
App::render('admin/users',compact('q','r','columns','actions'));
