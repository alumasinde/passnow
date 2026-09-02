<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
if(!$id){http_response_code(400);exit('Invalid visitor ID.');}

$errors=[];$visitor=[];$idTypes=[];$companies=[];
try{$p=Auth::api(App::api(),'GET','/api/v1/visitors/'.$id);$visitor=apiValue($p,'visitor',$p['data']??$p);if(!is_array($visitor))$visitor=[];}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to load visitor.';}
try{$p=Auth::api(App::api(),'GET','/api/v1/id-types');$idTypes=apiRows($p);}catch(Throwable){$errors[]='Unable to load ID types.';}
try{$p=Auth::api(App::api(),'GET','/api/v1/visitor-companies');$companies=apiRows($p);}catch(Throwable){}

$nullable=static fn(string $value):?string=>($value=trim($value))===''?null:$value;
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
 ];
 if($payload['first_name']==='')$errors[]='First name is required.';
 if($payload['last_name']==='')$errors[]='Last name is required.';
 if($payload['id_type_id']<1)$errors[]='ID type is required.';
 if($payload['email']!==null&&!filter_var($payload['email'],FILTER_VALIDATE_EMAIL))$errors[]='A valid email is required.';
 if(!$errors)try{
  Auth::api(App::api(),'PATCH','/api/v1/visitors/'.$id,$payload);
  flash('success','Visitor updated successfully.');
  redirect('visitor.php?id='.rawurlencode((string)$id));
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to update visitor right now.';}
 $visitor=array_merge($visitor,$payload);
}

App::render('visitors/edit',compact('id','visitor','idTypes','companies','errors'));