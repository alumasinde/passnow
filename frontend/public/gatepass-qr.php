<?php
declare(strict_types=1);
require_once __DIR__ . '/../app/App.php';

Auth::requireLogin();
$token=trim((string)($_GET['token']??''));
if($token===''||!preg_match('/^[a-f0-9]{32}$/i',$token)){http_response_code(400);exit;}
try{
    $response=Auth::apiBinary(App::api(),'GET','/api/v1/gatepasses/qr/image/'.rawurlencode($token));
    header('Content-Type: '.$response['content_type']);
    header('Cache-Control: no-store, private');
    echo $response['body'];
}catch(Throwable){http_response_code(404);}
