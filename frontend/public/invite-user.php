<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$errors=[];$roles=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/roles');$roles=apiRows($p);}catch(Throwable){}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=['email'=>trim((string)($_POST['email']??'')),'first_name'=>trim((string)($_POST['first_name']??'')),'last_name'=>trim((string)($_POST['last_name']??'')),'role_id'=>(int)($_POST['role_id']??0)];
 if(!filter_var($payload['email'],FILTER_VALIDATE_EMAIL))$errors[]='A valid email is required.';
 if($payload['first_name']==='')$errors[]='First name is required.';
 if($payload['last_name']==='')$errors[]='Last name is required.';
 if(!$errors)try{Auth::api(App::api(),'POST','/api/v1/users/invite',$payload);flash('success','Invitation created successfully.');redirect('users.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create invitation.';}
}
App::render('admin/invite-user',compact('errors','roles'));
