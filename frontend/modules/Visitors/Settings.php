<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$data=[];$errors=[];
try{
    $p=Auth::api(App::api(),'GET','/api/v1/settings/visitors');
    $data=apiValue($p,'settings',$p['data']??$p);
    if(!is_array($data))$data=[];
}catch(ApiException $e){$errors[]=$e->getMessage();}
catch(Throwable){$errors[]='Unable to load visitor settings.';}

if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $payload=['allow_pre_registration'=>isset($_POST['allow_pre_registration']) && (string)$_POST['allow_pre_registration']==='1'];
    try{
        $data=Auth::api(App::api(),'PUT','/api/v1/settings/visitors',$payload);
        flash('success','Visitor settings updated.');
        redirect('visitor-settings.php');
    }catch(ApiException $e){
        $errors[]=$e->getMessage();
        $data=array_merge($data,$payload);
    }catch(Throwable){
        $errors[]='Unable to save visitor settings.';
        $data=array_merge($data,$payload);
    }
}
App::render('admin/visitor-settings',compact('data','errors'));
