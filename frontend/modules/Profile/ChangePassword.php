<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$errors=[];
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $current=(string)($_POST['current_password']??'');$new=(string)($_POST['new_password']??'');$confirm=(string)($_POST['confirm_password']??'');
 if($current===''||$new===''||$confirm==='')$errors[]='All password fields are required.';
 if($new!==$confirm)$errors[]='New password and confirmation do not match.';
 if(strlen($new)<8)$errors[]='New password must be at least 8 characters.';
 if(!$errors)try{Auth::api(App::api(),'POST','/api/v1/auth/change-password',['current_password'=>$current,'new_password'=>$new]);$_SESSION['must_change_password']=false;flash('success','Password changed successfully.');redirect('profile.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to change password.';}
}
App::render('profile/change-password',compact('errors'));