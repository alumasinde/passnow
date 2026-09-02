<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid role ID.');}
$role=[];$errors=[];try{$p=Auth::api(App::api(),'GET','/api/v1/roles/'.$id);$role=apiValue($p,'role',$p['data']??$p);if(!is_array($role))$role=[];}catch(Throwable){$errors[]='Unable to load role.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);$name=trim((string)($_POST['name']??''));
 if($name==='')$errors[]='Role name is required.';
 if(!empty($role['is_system']))$errors[]='System roles cannot be renamed.';
 if(!$errors)try{Auth::api(App::api(),'PATCH','/api/v1/roles/'.$id,['name'=>$name]);flash('success','Role updated successfully.');redirect('roles.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to update role.';}
 $role['name']=$name;
}
App::render('admin/roles-edit',compact('id','role','errors'));