<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$errors=[];$idTypes=[];$companies=[];$preRegistrationEnabled=false;
try{$p=Auth::api(App::api(),'GET','/api/v1/id-types');$idTypes=apiRows($p);}catch(Throwable $e){$errors[]='Unable to load ID types.';}
try{$p=Auth::api(App::api(),'GET','/api/v1/visitor-companies');$companies=apiRows($p);}catch(Throwable){}
try{
 $settings=Auth::api(App::api(),'GET','/api/v1/settings/visitors');
 $preRegistrationEnabled=(bool)apiValue($settings,'allow_pre_registration',apiValue($settings,'visitors_allow_pre_registration',false));
}catch(Throwable){}

$nullable=static fn(string $value): ?string => ($value=trim($value))===''?null:$value;
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=[
  'first_name'=>trim((string)($_POST['first_name']??'')),
  'last_name'=>trim((string)($_POST['last_name']??'')),
  'id_type_id'=>(int)($_POST['id_type_id']??0),
  'id_number'=>$nullable((string)($_POST['id_number']??'')),
  'company_id'=>(int)($_POST['company_id']??0)?:null,
  'phone'=>$nullable((string)($_POST['phone']??'')),
  'email'=>$nullable((string)($_POST['email']??'')),
  'notes'=>$nullable((string)($_POST['notes']??'')),
  'pre_register'=>$preRegistrationEnabled && isset($_POST['pre_register']),
 ];
 if($payload['first_name']==='')$errors[]='First name is required.';
 if($payload['last_name']==='')$errors[]='Last name is required.';
 if($payload['id_type_id']<1)$errors[]='ID type is required.';
 if(!$errors)try{
  $created=Auth::api(App::api(),'POST','/api/v1/visitors',$payload);
  $id=apiValue($created,'id');flash('success',$payload['pre_register']?'Visitor pre-registered successfully.':'Visitor created successfully.');
  redirect($id?'visitor.php?id='.rawurlencode((string)$id):'visitors.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create visitor right now.';}
}
App::render('visitors/create',compact('errors','idTypes','companies','preRegistrationEnabled'));
