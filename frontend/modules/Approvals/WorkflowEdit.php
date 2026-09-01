<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);$item=[];$errors=[];
if($id)try{$p=Auth::api(App::api(),'GET','/api/v1/approval-workflows/'.$id);$item=apiValue($p,'workflow',$p['data']??$p);if(!is_array($item))$item=[];}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to load workflow.';}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $raw=trim((string)($_POST['steps_json']??'[]'));$steps=json_decode($raw,true);
 $payload=['name'=>trim((string)($_POST['name']??'')),'description'=>trim((string)($_POST['description']??'')),'steps'=>$steps];
 if($payload['name']==='')$errors[]='Workflow name is required.';
 if(!is_array($steps))$errors[]='Steps must be valid JSON.';
 if(!$errors)try{
  if($id)Auth::api(App::api(),'PATCH','/api/v1/approval-workflows/'.$id,$payload);else Auth::api(App::api(),'POST','/api/v1/approval-workflows',$payload);
  flash('success','Approval workflow saved.');redirect('approval-workflows.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save workflow.';}
}
$stepsJson=json_encode($item['steps']??[],JSON_PRETTY_PRINT|JSON_UNESCAPED_SLASHES);
App::render('admin/approval-workflow-edit',compact('id','item','errors','stepsJson'));
