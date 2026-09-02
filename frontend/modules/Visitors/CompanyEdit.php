<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);$item=[];$errors=[];
if($id){try{$p=Auth::api(App::api(),'GET','/api/v1/visitor-companies/'.$id);$item=apiValue($p,'item',$p['data']??$p);if(!is_array($item))$item=[];}catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load visitor company.';}}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $nullable=static fn(string $v):?string=>($v=trim($v))===''?null:$v;
 $payload=['name'=>trim((string)($_POST['name']??'')),'phone'=>$nullable((string)($_POST['phone']??'')),'email'=>$nullable((string)($_POST['email']??'')),'address'=>$nullable((string)($_POST['address']??'')),'active'=>isset($_POST['active'])];
 if($payload['name']==='')$errors[]='Company name is required.';
 if($payload['email']!==null&&!filter_var($payload['email'],FILTER_VALIDATE_EMAIL))$errors[]='Enter a valid email address.';
 if(!$errors)try{
  if($id){ConfigCrud::update('/api/v1/visitor-companies/'.$id,$payload);flash('success','Visitor company updated successfully.');}
  else{ConfigCrud::create('/api/v1/visitor-companies',$payload);flash('success','Visitor company created successfully.');}
  redirect('visitor-companies.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this visitor company.';}
 $item=array_merge($item,$payload);
}
App::render('admin/visitor-companies-edit',compact('id','item','errors'));