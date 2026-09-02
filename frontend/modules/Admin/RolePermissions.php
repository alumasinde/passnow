<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid role ID.');}
$role=[];$permissions=[];$selected=[];$error=null;
try{$p=Auth::api(App::api(),'GET','/api/v1/roles');foreach(apiRows($p) as $row)if((int)($row['id']??0)===$id)$role=$row;}catch(Throwable){}
try{$p=Auth::api(App::api(),'GET','/api/v1/permissions');$permissions=apiRows($p);}catch(ApiException $e){$error=$e->getMessage();}catch(Throwable){$error='Unable to load permissions.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $ids=[];foreach((array)($_POST['permissions']??[]) as $v){$v=filter_var($v,FILTER_VALIDATE_INT);if($v)$ids[]=$v;}
 try{Auth::api(App::api(),'PUT','/api/v1/roles/'.$id.'/permissions',['permission_ids'=>$ids]);flash('success','Role permissions updated.');redirect('role-permissions.php?id='.$id);}catch(ApiException $e){$error=$e->getMessage();}catch(Throwable){$error='Unable to update role permissions.';}
}
$selected=array_map('intval',(array)($role['permission_ids']??$role['permissions']??[]));
App::render('admin/role-permissions',compact('id','role','permissions','selected','error'));
