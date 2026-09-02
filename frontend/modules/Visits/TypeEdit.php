<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);$item=[];$errors=[];
if($id){try{$p=Auth::api(App::api(),'GET','/api/v1/visit-types/'.$id);$item=apiValue($p,'item',$p['data']??$p);if(!is_array($item))$item=[];}catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load visit type.';}}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $description=trim((string)($_POST['description']??''));
 $payload=['name'=>trim((string)($_POST['name']??'')),'code'=>strtoupper(trim((string)($_POST['code']??''))),'description'=>$description,'active'=>!empty($_POST['active'])];
 if($payload['name']==='')$errors[]='Visit type name is required.';
 if($payload['code']==='')$errors[]='Visit type code is required.';
 if(!$errors)try{
  if($id){ConfigCrud::update('/api/v1/visit-types/'.$id,$payload);flash('success','Visit type updated successfully.');}
  else{ConfigCrud::create('/api/v1/visit-types',$payload);flash('success','Visit type created successfully.');}
  redirect('visit-types.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this visit type.';}
 $item=array_merge($item,$payload);
}
App::render('admin/visit-types-edit',compact('id','item','errors'));
