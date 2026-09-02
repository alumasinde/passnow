<?php
declare(strict_types=1); require_once __DIR__.'/../../app/App.php'; Auth::requireLogin();
$errors=[];$types=[];$departments=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/visit-types');$types=apiRows($p);}catch(Throwable){}
try{$p=Auth::api(App::api(),'GET','/api/v1/departments');$departments=apiRows($p);}catch(Throwable){}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=['visitor_id'=>(int)($_POST['visitor_id']??0),'visit_type_id'=>(int)($_POST['visit_type_id']??0),'department_id'=>(int)($_POST['department_id']??0),'purpose'=>trim((string)($_POST['purpose']??'')),'expected_time'=>trim((string)($_POST['expected_time']??''))];
 if($payload['visitor_id']<1)$errors[]='Visitor is required.'; if($payload['visit_type_id']<1)$errors[]='Visit type is required.'; if($payload['purpose']==='')$errors[]='Purpose is required.';
 if(!$errors)try{$p=Auth::api(App::api(),'POST','/api/v1/visits',$payload);$id=apiValue($p,'id');flash('success','Visit created successfully.');redirect($id?'visit.php?id='.rawurlencode((string)$id):'visits.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create visit.';}
}
App::render('visits/create',compact('errors','types','departments'));
