<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$data=[];$errors=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/settings/visitors');$data=apiValue($p,'settings',$p['data']??$p);if(!is_array($data))$data=[];}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to load settings.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);$payload=[];
 foreach($_POST as $k=>$v)if($k!=='_csrf')$payload[$k]=is_string($v)?trim($v):$v;
 try{Auth::api(App::api(),'PUT','/api/v1/settings/visitors',$payload);flash('success','Settings updated.');redirect('visitor-settings.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save settings.';}
}
App::render('admin/visitor-settings',compact('data','errors'));
