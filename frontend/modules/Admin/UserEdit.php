<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid user ID.');}
$errors=[];$user=[];$roles=[];$departments=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/users/'.$id);$user=apiValue($p,'user',$p['data']??$p);if(!is_array($user))$user=[];}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to load user.';}
try{$roles=apiRows(Auth::api(App::api(),'GET','/api/v1/roles'));}catch(Throwable){$errors[]='Unable to load roles.';}
try{$departments=apiRows(Auth::api(App::api(),'GET','/api/v1/departments'));}catch(Throwable){$errors[]='Unable to load departments.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);$membershipID=(int)($user['membership_id']??0);$departmentRaw=trim((string)($_POST['department_id']??''));
 $payload=['role_id'=>(int)($_POST['role_id']??0),'status'=>trim((string)($_POST['status']??'')),'clear_department'=>$departmentRaw===''];if($departmentRaw!=='')$payload['department_id']=(int)$departmentRaw;
 if($membershipID<1)$errors[]='User membership was not found.';if($payload['role_id']<1)$errors[]='Role is required.';if(!in_array($payload['status'],['active','invited','disabled'],true))$errors[]='A valid status is required.';
 if(!$errors)try{Auth::api(App::api(),'PATCH','/api/v1/users/memberships/'.$membershipID,$payload);flash('success','User updated successfully.');redirect('user.php?id='.rawurlencode((string)$id));}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to update user.';}
 $user=array_merge($user,$payload);
}
App::render('admin/user-edit',compact('id','user','roles','departments','errors'));