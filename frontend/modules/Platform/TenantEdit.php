<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requirePlatform();
$id=(int)($_GET['id']??0);if($id<1){redirect('platform/tenants');}
$error='';$tenant=[];
try{
 if(requestMethod()==='POST'){Csrf::requireValid($_POST['_csrf']??null);$action=(string)($_POST['action']??'update');
  if($action==='update') Auth::platformApi(App::api(),'PATCH','/api/v1/platform/tenants/'.$id,['name'=>trim((string)$_POST['name']),'slug'=>strtolower(trim((string)$_POST['slug']))]);
  elseif($action==='add_domain') Auth::platformApi(App::api(),'POST','/api/v1/platform/tenants/'.$id.'/domains',['domain'=>trim((string)$_POST['domain'])]);
  elseif($action==='primary') Auth::platformApi(App::api(),'PATCH','/api/v1/platform/tenants/'.$id.'/domains/'.(int)$_POST['domain_id'].'/primary',[]);
  elseif($action==='delete_domain') Auth::platformApi(App::api(),'DELETE','/api/v1/platform/tenants/'.$id.'/domains/'.(int)$_POST['domain_id']);
  flash('success','Tenant updated.');redirect('platform/tenant-edit?id='.$id);
 }
 $tenant=Auth::platformApi(App::api(),'GET','/api/v1/platform/tenants/'.$id);
}catch(Throwable $e){$error=$e->getMessage();try{$tenant=Auth::platformApi(App::api(),'GET','/api/v1/platform/tenants/'.$id);}catch(Throwable $_){}}
App::render('admin/tenants/edit',['tenant'=>$tenant,'error'=>$error],'platform');