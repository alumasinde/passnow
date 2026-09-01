<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);if(!$id){http_response_code(400);exit('Invalid visit ID.');}
$visit=[];$error=null;try{$p=Auth::api(App::api(),'GET','/api/v1/visits/'.$id);$visit=apiValue($p,'visit',$p['data']??$p);if(!is_array($visit))$visit=[];}catch(ApiException $e){$error=$e->getMessage();}catch(Throwable){$error='Unable to load visit.';}
App::render('visits/show',compact('id','visit','error'));
