<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
if(!$id){http_response_code(400);exit('Invalid approval ID.');}

$approval=[];$error=null;
try{
 $payload=Auth::api(App::api(),'GET','/api/v1/gatepasses/'.$id);
 $approval=apiValue($payload,'gatepass',$payload['data']??$payload);
 if(!is_array($approval))$approval=[];
}catch(ApiException $e){$error=$e->getMessage();}
 catch(Throwable){$error='Unable to load the approval record.';}

$steps=[];
try{
 $payload=Auth::api(App::api(),'GET','/api/v1/gatepasses/'.$id.'/approvals');
 $steps=apiRows($payload);
}catch(Throwable){}

App::render('approvals/show',compact('id','approval','steps','error'));
