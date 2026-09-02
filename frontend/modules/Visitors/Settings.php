<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$isAjax=strtolower((string)($_SERVER['HTTP_X_REQUESTED_WITH']??''))==='xmlhttprequest';
$json=static function(int $status,array $payload):never{http_response_code($status);header('Content-Type: application/json; charset=utf-8');echo json_encode($payload,JSON_UNESCAPED_SLASHES);exit;};

$data=[];$errors=[];
try{
    $response=Auth::api(App::api(),'GET','/api/v1/settings/visitors');
    $data=is_array($response['data']??null)?$response['data']:$response;
    if(!is_array($data))$data=[];
}catch(ApiException $e){$errors[]=$e->getMessage();}
catch(Throwable){$errors[]='Unable to load visitor settings.';}

if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $payload=['allow_pre_registration'=>isset($_POST['allow_pre_registration']) && (string)$_POST['allow_pre_registration']==='1'];
    try{
        $response=Auth::api(App::api(),'PUT','/api/v1/settings/visitors',$payload);
        $data=is_array($response['data']??null)?$response['data']:$response;
        if(!is_array($data))$data=$payload;
        if($isAjax)$json(200,['ok'=>true,'message'=>'Visitor settings saved.','data'=>$data]);
        flash('success','Visitor settings updated.');redirect('visitor-settings.php');
    }catch(ApiException $e){$errors[]=$e->getMessage();$data=array_merge($data,$payload);}
    catch(Throwable){$errors[]='Unable to save visitor settings.';$data=array_merge($data,$payload);}
    if($isAjax)$json(422,['ok'=>false,'message'=>$errors[0]??'Unable to save visitor settings.','errors'=>$errors]);
}
App::render('admin/visitor-settings',compact('data','errors'));
