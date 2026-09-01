<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requirePlatform();
$id=(int)($_GET['id']??0);if($id<1){redirect('platform/tenants');}
$error='';$tenant=[];
try{$tenant=Auth::platformApi(App::api(),'GET','/api/v1/platform/tenants/'.$id);}catch(Throwable $e){$error=$e->getMessage();}
App::render('admin/tenants/view',['tenant'=>$tenant,'error'=>$error],'platform');