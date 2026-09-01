<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';
Auth::requireLogin();

if(requestMethod()!=='POST'){http_response_code(405);exit('Method not allowed.');}
Csrf::requireValid($_POST['_csrf']??null);

$gatepassId=filter_var($_POST['gatepass_id']??null,FILTER_VALIDATE_INT);
$stepId=filter_var($_POST['step_id']??null,FILTER_VALIDATE_INT);
$decision=(string)($_POST['decision']??'');
$comment=trim((string)($_POST['comment']??''));

if(!$gatepassId||!$stepId||!in_array($decision,['approve','reject'],true)){http_response_code(400);exit('Invalid approval decision.');}

$routes=[
 'approve'=>'/api/v1/gatepasses/%d/approvals/%d/approve',
 'reject'=>'/api/v1/gatepasses/%d/approvals/%d/reject'
];
try{
 Auth::api(App::api(),'POST',sprintf($routes[$decision],$gatepassId,$stepId),$comment!==''?['comment'=>$comment]:[]);
 flash('success',$decision==='approve'?'Approval recorded successfully.':'Gatepass rejected.');
}catch(ApiException $e){flash('danger',$e->getMessage());}
 catch(Throwable){flash('danger','The approval decision could not be completed.');}
redirect('approval.php?id='.rawurlencode((string)$gatepassId));
