<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
$operation=(string)($_GET['operation']??'');
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $id=filter_var($_POST['id']??null,FILTER_VALIDATE_INT);
 $operation=(string)($_POST['operation']??'');
}
if(!$id||!in_array($operation,['check-out','check-in'],true)){http_response_code(400);exit('Invalid operation.');}

if(requestMethod()==='GET'){
 $payload=Auth::api(App::api(),'GET','/api/v1/gatepasses/'.$id);
 $gatepass=apiValue($payload,'gatepass',$payload['data']??$payload);
 App::render('operations/confirm',compact('id','operation','gatepass'));
 exit;
}

$routes=['check-out'=>'/api/v1/gatepasses/%d/check-out','check-in'=>'/api/v1/gatepasses/%d/check-in'];
try{
 Auth::api(App::api(),'POST',sprintf($routes[$operation],$id));
 flash('success',$operation==='check-out'?'Gatepass checked out successfully.':'Gatepass checked in successfully.');
}catch(ApiException $e){flash('danger',$e->getMessage());}
 catch(Throwable){flash('danger','The gate operation could not be completed.');}
redirect('gatepass.php?id='.rawurlencode((string)$id));
