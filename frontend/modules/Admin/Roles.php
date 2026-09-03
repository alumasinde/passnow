<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);$name=trim((string)($_POST['name']??''));
 try{if($name==='')throw new RuntimeException('Role name is required.');$role=Auth::api(App::api(),'POST','/api/v1/roles',['name'=>$name]);$payload=['ok'=>true,'message'=>'Role created successfully.','role'=>$role];if(strtolower((string)($_SERVER['HTTP_X_REQUESTED_WITH']??''))==='xmlhttprequest'){header('Content-Type: application/json');echo json_encode($payload);exit;}flash('success','Role created successfully.');redirect('roles.php');}catch(Throwable $e){if(strtolower((string)($_SERVER['HTTP_X_REQUESTED_WITH']??''))==='xmlhttprequest'){http_response_code(422);header('Content-Type: application/json');echo json_encode(['ok'=>false,'message'=>$e->getMessage()]);exit;}flash('error',$e->getMessage());redirect('roles.php');}
}
$q=new ListQuery([]);$r=ResourcePage::list('/api/v1/roles',$q,[],static function(array $rows):array{
 foreach($rows as &$row){$row['description']=!empty($row['is_system'])?'System role':'Custom tenant role';$row['user_count']=$row['user_count']??'—';}
 unset($row);return $rows;
});
$columns=[['key'=>'name','label'=>'Role'],['key'=>'description','label'=>'Type'],['key'=>'user_count','label'=>'Users']];
$canCreate=Auth::can('role.create');
$actions=[['label'=>'Edit role','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('roles-edit.php?id='.rawurlencode((string)($row['id']??'')))],['label'=>'Permissions','icon'=>'fa-key','class'=>'btn-secondary','href'=>fn($row)=>url('role-permissions.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/roles',compact('q','r','columns','actions','canCreate'));
