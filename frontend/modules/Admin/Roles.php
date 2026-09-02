<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=ResourcePage::list('/api/v1/roles',$q,[],static function(array $rows):array{
 foreach($rows as &$row){$row['description']=!empty($row['is_system'])?'System role':'Custom tenant role';$row['user_count']=$row['user_count']??'—';}
 unset($row);return $rows;
});
$columns=[['key'=>'name','label'=>'Role'],['key'=>'description','label'=>'Type'],['key'=>'user_count','label'=>'Users']];
$actions=[['label'=>'Permissions','icon'=>'fa-key','class'=>'btn-secondary','href'=>fn($row)=>url('role-permissions.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/roles',compact('q','r','columns','actions'));
