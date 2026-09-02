<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);$item=[];$errors=[];
if($id){try{$p=Auth::api(App::api(),'GET','/api/v1/gatepass-types/'.$id);$item=apiValue($p,'item',$p['data']??$p);if(!is_array($item))$item=[];}catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load record.';}}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=[];
 $payload['name']=trim((string)($_POST['name']??''));$payload['description']=trim((string)($_POST['description']??''));
 if(!$errors)try{
  if($id){ConfigCrud::update('/api/v1/gatepass-types/'.$id,$payload);flash('success','Updated successfully.');}
  else{ConfigCrud::create('/api/v1/gatepass-types',$payload);flash('success','Created successfully.');}
  redirect('gatepass-types.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this record.';}
}
App::render('admin/gatepass-types-edit',compact('id','item','errors'));
