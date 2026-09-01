<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid user ID.');}
$user=[];$error=null;try{$p=Auth::api(App::api(),'GET','/api/v1/users/'.$id);$user=apiValue($p,'user',$p['data']??$p);if(!is_array($user))$user=[];}catch(ApiException $e){$error=$e->getMessage();}catch(Throwable){$error='Unable to load user.';}
App::render('admin/user',compact('id','user','error'));
