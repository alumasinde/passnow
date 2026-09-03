<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();
if(requestMethod()!=='POST'){http_response_code(405);exit('Method not allowed.');}
Csrf::requireValid($_POST['_csrf']??null);
$id=filter_var($_POST['id']??null,FILTER_VALIDATE_INT);$op=(string)($_POST['operation']??'');
if(!$id||!in_array($op,['check-in','check-out','cancel'],true)){http_response_code(400);exit('Invalid operation.');}
try{
 if($op==='cancel'){
  $reason=trim((string)($_POST['reason']??''));if($reason===''){flash('danger','A cancellation reason is required.');redirect('visit.php?id='.rawurlencode((string)$id));}
  Auth::api(App::api(),'POST','/api/v1/visits/'.$id.'/cancel',['reason'=>$reason]);flash('success','Visit cancelled.');
 }else{
  $gateID=filter_var($_POST['gate_id']??null,FILTER_VALIDATE_INT);
  $deviceID=filter_var($_POST['device_id']??null,FILTER_VALIDATE_INT);
  if(!$gateID){flash('danger','Select a gate before performing this operation.');redirect('visit.php?id='.rawurlencode((string)$id));}
  $payload=['gate_id'=>(int)$gateID];
  if($deviceID){$payload['device_id']=(int)$deviceID;}
  $notes=trim((string)($_POST['notes']??''));if($notes!==''){$payload['notes']=$notes;}
  Auth::api(App::api(),'POST','/api/v1/visits/'.$id.'/'.$op,$payload);flash('success',$op==='check-in'?'Visit checked in and badge issued.':'Visit checked out.');
 }
}catch(ApiException $e){flash('danger',$e->getMessage());}catch(Throwable){flash('danger','Visit operation failed.');}
redirect('visit.php?id='.rawurlencode((string)$id));
