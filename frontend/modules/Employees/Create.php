<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$errors=[];$departments=[];try{$p=Auth::api(App::api(),'GET','/api/v1/departments');$departments=apiRows($p);}catch(Throwable){}
if(requestMethod()==='POST'){
 Csrf::requireValid($_POST['_csrf']??null);
 $payload=['employee_number'=>trim((string)($_POST['employee_number']??'')),'first_name'=>trim((string)($_POST['first_name']??'')),'last_name'=>trim((string)($_POST['last_name']??'')),'phone'=>trim((string)($_POST['phone']??'')),'email'=>trim((string)($_POST['email']??'')),'department_id'=>(int)($_POST['department_id']??0)];
 foreach(['employee_number'=>'Employee number is required.','first_name'=>'First name is required.','last_name'=>'Last name is required.'] as $k=>$msg)if($payload[$k]==='')$errors[]=$msg;
 if(!$errors)try{$p=Auth::api(App::api(),'POST','/api/v1/employees',$payload);$id=apiValue($p,'id');flash('success','Employee created successfully.');redirect($id?'employee.php?id='.rawurlencode((string)$id):'employees.php');}catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create employee.';}
}
App::render('employees/create',compact('errors','departments'));
