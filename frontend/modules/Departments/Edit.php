<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);$item=[];$errors=[];
if($id){try{$p=Auth::api(App::api(),'GET','/api/v1/departments/'.$id);$item=apiValue($p,'item',$p['data']??$p);if(!is_array($item))$item=[];}catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load record.';}}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 error_log('PASSNOW DEPARTMENT FORM: method='.requestMethod().' post_keys='.implode(',', array_keys($_POST)));
 $payload=[];
 $payload['name']=trim((string)($_POST['name']??''));
 $payload['code']=strtoupper(trim((string)($_POST['code']??'')));
 $payload['active']=isset($_POST['active']);
 $payload['active']=isset($_POST['active']);
 if($payload['name']===''||$payload['code']==='')$errors[]='Name and code are required.';
 if(!$errors)try{
  if($id){ConfigCrud::update('/api/v1/departments/'.$id,$payload);flash('success','Updated successfully.');}
  else{ConfigCrud::create('/api/v1/departments',$payload);flash('success','Created successfully.');}
  redirect('departments.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this record.';}
}
if (requestMethod()==='POST') $item=array_merge($item,['name'=>$payload['name']??'','code'=>$payload['code']??'','active'=>$payload['active']??false]);
App::render('admin/departments-edit',compact('id','item','errors'));
