<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
if(requestMethod()!=='POST'){http_response_code(405);exit('Method not allowed.');}
Csrf::requireValid($_POST['_csrf']??null);$id=filter_var($_POST['id']??null,FILTER_VALIDATE_INT);$op=(string)($_POST['operation']??'');
if(!$id||!in_array($op,['check-in','check-out'],true)){http_response_code(400);exit('Invalid operation.');}
$route=['check-in'=>'/api/v1/visits/%d/check-in','check-out'=>'/api/v1/visits/%d/check-out'];
try{Auth::api(App::api(),'POST',sprintf($route[$op],$id));flash('success',$op==='check-in'?'Visit checked in.':'Visit checked out.');}catch(ApiException $e){flash('danger',$e->getMessage());}catch(Throwable){flash('danger','Visit operation failed.');}
redirect('visit.php?id='.rawurlencode((string)$id));
