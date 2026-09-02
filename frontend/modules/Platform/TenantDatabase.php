<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requirePlatform();
$id=(int)($_GET['id']??0);if($id<1){redirect('platform/tenants');}
$error='';$tenant=[];$health=[];
try{
 if(requestMethod()==='POST'){
  Csrf::requireValid($_POST['_csrf']??null);
  $action=(string)($_POST['action']??'');
  if($action==='migrate'){
   Auth::platformApi(App::api(),'POST','/api/v1/platform/tenants/'.$id.'/run-migrations',[]);
   flash('success','Tenant migrations completed successfully.');
   redirect('platform/tenant-database?id='.$id);
  }
 }
 $tenant=Auth::platformApi(App::api(),'GET','/api/v1/platform/tenants/'.$id);
 $health=Auth::platformApi(App::api(),'GET','/api/v1/platform/tenants/'.$id.'/database-health');
}catch(Throwable $e){$error=$e->getMessage();}
App::render('admin/tenants/database',compact('id','tenant','health','error'),'platform');
