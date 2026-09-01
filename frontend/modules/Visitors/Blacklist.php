<?php
declare(strict_types=1);
require_once __DIR__.'/../app/App.php'; Auth::requireLogin();
if(requestMethod()!=='POST'){http_response_code(405);exit('Method not allowed.');}
Csrf::requireValid($_POST['_csrf']??null);$id=filter_var($_POST['id']??null,FILTER_VALIDATE_INT);
if(!$id){http_response_code(400);exit('Invalid visitor ID.');}
try{Auth::api(App::api(),'POST','/api/v1/visitors/'.$id.'/blacklist');flash('success','Visitor blacklist status updated.');}
catch(ApiException $e){flash('danger',$e->getMessage());}catch(Throwable){flash('danger','Unable to update visitor status.');}
redirect('visitor.php?id='.rawurlencode((string)$id));
