<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
if(!$id){http_response_code(400);exit('Invalid visitor ID.');}
$visitor=[];$error=null;
try{$p=Auth::api(App::api(),'GET','/api/v1/visitors/'.$id);$visitor=apiValue($p,'visitor',$p['data']??$p);if(!is_array($visitor))$visitor=[];}
catch(ApiException $e){$error=$e->getMessage();}catch(Throwable){$error='Unable to load visitor.';}
App::render('visitors/show',compact('id','visitor','error'));
