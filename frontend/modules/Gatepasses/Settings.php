<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$data=[];$errors=[];
try{
    $response=Auth::api(App::api(),'GET','/api/v1/settings/gatepass');
    // The Go handler returns the settings object directly. Preserve database
    // values exactly; only unwrap known envelopes if the API changes later.
    $data=is_array($response['data']??null)?$response['data']:$response;
    if(!is_array($data))$data=[];
    $data['number_prefix']=trim((string)($data['number_prefix']??''));
    $data['number_use_year']=array_key_exists('number_use_year',$data)
        ? filter_var($data['number_use_year'],FILTER_VALIDATE_BOOLEAN,FILTER_NULL_ON_FAILURE) ?? false
        : null;
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
            $response=Auth::api(App::api(),'PUT','/api/v1/settings/gatepass',$payload);
            $data=is_array($response['data']??null)?$response['data']:$response;
            if(!is_array($data))$data=$payload;
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
