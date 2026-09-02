<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$errors=[];$success=false;$profile=Auth::user();
try{$p=Auth::api(App::api(),'GET','/api/v1/auth/me');if(is_array($p))$profile=$p;}catch(Throwable){}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=['first_name'=>trim((string)($_POST['first_name']??'')),'last_name'=>trim((string)($_POST['last_name']??''))];
 if($payload['first_name']===''||$payload['last_name']==='')$errors[]='First name and last name are required.';
 if(!$errors)try{$profile=Auth::api(App::api(),'PATCH','/api/v1/auth/me',$payload);$_SESSION['user']=is_array($profile)?$profile:$_SESSION['user'];flash('success','Profile updated successfully.');redirect('profile.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to update profile.';}
}
App::render('profile/index',compact('profile','errors'));