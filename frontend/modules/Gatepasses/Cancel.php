<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_POST,'id',FILTER_VALIDATE_INT);
$reason=trim((string)($_POST['reason']??''));
if(requestMethod()!=='POST'||!$id){http_response_code(400);exit('Invalid cancellation request.');}
Csrf::requireValid($_POST['_csrf']??null);
if($reason===''){flash('danger','A cancellation reason is required.');redirect('gatepass.php?id='.rawurlencode((string)$id));}
try{
    Auth::api(App::api(),'POST','/api/v1/gatepasses/'.rawurlencode((string)$id).'/cancel',['reason'=>$reason]);
    flash('success','Gatepass cancelled successfully.');
}catch(ApiException $e){flash('danger',$e->getMessage());}catch(Throwable){flash('danger','Unable to cancel this gatepass.');}
redirect('gatepass.php?id='.rawurlencode((string)$id));
