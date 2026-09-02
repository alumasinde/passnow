<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$data=[];$errors=[];
try{
    $p=Auth::api(App::api(),'GET','/api/v1/settings/gatepass');
    $data=apiValue($p,'settings',$p['data']??$p);
    if(!is_array($data))$data=[];
}catch(ApiException $e){$errors[]=$e->getMessage();}
catch(Throwable){$errors[]='Unable to load settings.';}

if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $payload=[
        'number_prefix'=>strtoupper(trim((string)($_POST['number_prefix']??''))),
        'number_use_year'=>isset($_POST['number_use_year']) && (string)$_POST['number_use_year']==='1',
    ];
    if($payload['number_prefix']===''){
        $errors[]='Pass number prefix is required.';
        $data=array_merge($data,$payload);
    }else{
        try{
            $data=Auth::api(App::api(),'PUT','/api/v1/settings/gatepass',$payload);
            flash('success','Gatepass settings updated.');
            redirect('gatepass-settings.php');
        }catch(ApiException $e){
            $errors[]=$e->getMessage();
            $data=array_merge($data,$payload);
        }catch(Throwable){
            $errors[]='Unable to save gatepass settings.';
            $data=array_merge($data,$payload);
        }
    }
}
App::render('admin/gatepass-settings',compact('data','errors'));
