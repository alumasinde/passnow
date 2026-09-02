<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$token=trim((string)($_GET['token']??''));
$record=[];$error=null;

if($token!==''){
 try{
  $payload=Auth::api(App::api(),'GET','/api/v1/gatepasses/qr/'.rawurlencode($token));
  $record=apiValue($payload,'gatepass',$payload['data']??$payload);
  if(!is_array($record))$record=[];
 }catch(ApiException $e){$error=$e->getMessage();}
  catch(Throwable){$error='Unable to look up the gatepass.';}
}
App::render('operations/index',compact('token','record','error'));
