<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'user_id',FILTER_VALIDATE_INT) ?: filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid user ID.');}
$errors=[];$user=[];$roles=[];$departments=[];$currentUser=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/auth/me');$currentUser=apiValue($p,'user',$p['data']??$p);if(!is_array($currentUser))$currentUser=[];}catch(Throwable){}
try{$p=Auth::api(App::api(),'GET','/api/v1/users/'.$id);$user=apiValue($p,'user',$p['data']??$p);if(!is_array($user))$user=[];}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to load user.';}
$isSelf=((int)($currentUser['id']??0)===$id);
try{$roles=apiRows(Auth::api(App::api(),'GET','/api/v1/roles'));}catch(Throwable){$errors[]='Unable to load roles.';}
try{$departments=apiRows(Auth::api(App::api(),'GET','/api/v1/departments'));}catch(Throwable){$errors[]='Unable to load departments.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);$membershipID=(int)($user['membership_id']??0);$departmentRaw=trim((string)($_POST['department_id']??''));
 $payload=[];$existingDepartment=$user['department_id']??null;$newDepartment=$departmentRaw===''?null:(int)$departmentRaw;
 if((string)($existingDepartment??'')!==(string)($newDepartment??'')){if($newDepartment===null)$payload['clear_department']=true;else $payload['department_id']=$newDepartment;}
 if(!$isSelf){
   $roleID=(int)($_POST['role_id']??0);$status=trim((string)($_POST['status']??''));
   if($roleID<1)$errors[]='Role is required.';if(!in_array($status,['active','invited','disabled'],true))$errors[]='A valid status is required.';
   if(!$errors && $roleID!==(int)($user['role_id']??0))$payload['role_id']=$roleID;
   if(!$errors && $status!==(string)($user['status']??''))$payload['status']=$status;
 }
 if($membershipID<1)$errors[]='User membership was not found.';
 if(!$errors)try{
   if($payload)Auth::api(App::api(),'PATCH','/api/v1/users/memberships/'.$membershipID,$payload);
   flash('success',$payload?'User updated successfully.':'No changes were needed.');
   redirect('admin/users/view?id='.rawurlencode((string)$id));
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to update user.';}
 $user=array_merge($user,$payload);if(!empty($payload['clear_department']))$user['department_id']=null;
}
App::render('admin/user-edit',compact('id','user','roles','departments','errors','isSelf'));