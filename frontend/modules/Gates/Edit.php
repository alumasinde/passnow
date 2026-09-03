<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
$item=[];$errors=[];
if($id){
    try{$p=Auth::api(App::api(),'GET','/api/v1/gates/'.$id);$item=apiValue($p,'item',$p['data']??$p);if(!is_array($item))$item=[];}
    catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load gate.';}
}
if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $payload=[
        'code'=>strtoupper(trim((string)($_POST['code']??''))),
        'name'=>trim((string)($_POST['name']??'')),
        'description'=>trim((string)($_POST['description']??'')) ?: null,
        'location'=>trim((string)($_POST['location']??'')) ?: null,
        'allows_entry'=>!empty($_POST['allows_entry']),
        'allows_exit'=>!empty($_POST['allows_exit']),
        'is_default'=>!empty($_POST['is_default']),
        'active'=>!empty($_POST['active']),
    ];
    if($payload['code']==='')$errors[]='Gate code is required.';
    if($payload['name']==='')$errors[]='Gate name is required.';
    if(!$payload['allows_entry']&&!$payload['allows_exit'])$errors[]='A gate must allow entry, exit, or both.';
    if(!$errors)try{
        if($id){ConfigCrud::update('/api/v1/gates/'.$id,$payload);flash('success','Gate updated successfully.');}
        else{ConfigCrud::create('/api/v1/gates',$payload);flash('success','Gate created successfully.');}
        redirect('gates.php');
    }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this gate.';}
    $item=array_merge($item,$payload);
}
App::render('admin/gates-edit',compact('id','item','errors'));