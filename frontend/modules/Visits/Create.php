<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();
$errors=[];$types=[];$departments=[];$visitors=[];$employees=[];$currentDepartmentID=null;
try{$p=Auth::api(App::api(),'GET','/api/v1/visit-types');$types=apiRows($p);}catch(Throwable $e){$errors[]='Unable to load visit types.';}
try{$p=Auth::api(App::api(),'GET','/api/v1/departments');$departments=apiRows($p);}catch(Throwable){}
try{$me=Auth::api(App::api(),'GET','/api/v1/auth/me');$me=apiValue($me,'user',$me['data']??$me);$currentDepartmentID=is_array($me)&&isset($me['department_id'])?(int)$me['department_id']:null;}catch(Throwable){}
try{$p=Auth::api(App::api(),'GET','/api/v1/visitors?limit=100');$visitors=apiRows($p);}catch(Throwable){}
try{$p=Auth::api(App::api(),'GET','/api/v1/employees?limit=100');$employees=apiRows($p);}catch(ApiException $e){$errors[]='Unable to load hosts: '.$e->getMessage();}catch(Throwable){$errors[]='Unable to load hosts.';}

$nullableInt=static fn($value): ?int => ((int)$value)>0?(int)$value:null;
$nullableString=static fn($value): ?string => (($v=trim((string)$value))==='')?null:$v;
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $expected=$nullableString($_POST['expected_time']??'');$expectedValue=null;
 if($expected!==null){try{$expectedValue=(new DateTimeImmutable($expected))->format(DATE_ATOM);}catch(Throwable){$errors[]='Expected time is invalid.';}}
 $payload=[
  'visitor_id'=>(int)($_POST['visitor_id']??0),
  'visit_type_id'=>$nullableInt($_POST['visit_type_id']??0),
  'department_id'=>$nullableInt($_POST['department_id']??0),
  'host_name'=>$nullableString($_POST['host_name']??''),
  'purpose'=>$nullableString($_POST['purpose']??''),
  'expected_time'=>$expectedValue,
  'check_in_now'=>isset($_POST['check_in_now']),
 ];
 if($payload['visitor_id']<1)$errors[]='Visitor is required.';
 if(!$payload['visit_type_id'])$errors[]='Visit type is required.';
 if(!$payload['purpose'])$errors[]='Purpose is required.';
 if(!$errors)try{
  $p=Auth::api(App::api(),'POST','/api/v1/visits',$payload);$id=apiValue($p,'id');
  flash('success',$payload['check_in_now']?'Visit created and checked in.':'Visit created successfully.');
  redirect($id?'visit.php?id='.rawurlencode((string)$id):'visits.php');
 }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create visit.';}
}
$preselectedVisitor=(int)($_GET['visitor_id']??oldOr('visitor_id'));
App::render('visits/create',compact('errors','types','departments','visitors','employees','preselectedVisitor','currentDepartmentID'));
