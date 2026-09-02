<?php
declare(strict_types=1);
require_once __DIR__ . '/../../app/App.php';
Auth::requireLogin();

$errors=[]; $types=[]; $visitors=[]; $departments=[]; $currentDepartmentID=null;
try { $types=apiRows(Auth::api(App::api(),'GET','/api/v1/gatepass-types')); } catch(Throwable){ $errors[]='Unable to load gatepass types.'; }
try { $visitors=apiRows(Auth::api(App::api(),'GET','/api/v1/visitors?limit=200')); } catch(Throwable){}
try { $departments=apiRows(Auth::api(App::api(),'GET','/api/v1/departments?limit=200')); } catch(Throwable){}
try { $me=Auth::api(App::api(),'GET','/api/v1/auth/me'); $me=apiValue($me,'user',$me['data']??$me); $currentDepartmentID=is_array($me)&&isset($me['department_id'])?(int)$me['department_id']:null; } catch(Throwable){}

$nullableInt=static fn($v): ?int => ((int)$v)>0?(int)$v:null;
$nullableString=static fn($v): ?string => (($s=trim((string)$v))==='')?null:$s;
if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $expected=null;
    if(($raw=trim((string)($_POST['expected_return_at']??'')))!==''){
        try{$expected=(new DateTimeImmutable($raw))->format(DATE_ATOM);}catch(Throwable){$errors[]='Expected return time is invalid.';}
    }
    $items=[];
    foreach(($_POST['items']??[]) as $item){
        if(!is_array($item)) continue;
        $name=trim((string)($item['name']??''));
        if($name==='') continue;
        $items[]=[
            'name'=>$name,
            'description'=>$nullableString($item['description']??''),
            'category'=>$nullableString($item['category']??''),
            'quantity'=>(float)($item['quantity']??0),
            'unit'=>$nullableString($item['unit']??''),
            'serial_number'=>$nullableString($item['serial_number']??''),
            'asset_number'=>$nullableString($item['asset_number']??''),
            'condition'=>$nullableString($item['condition']??''),
            'direction'=>trim((string)($item['direction']??'leaving')),
        ];
    }
    $returnable=isset($_POST['is_returnable']);
    $payload=[
        'gatepass_type_id'=>(int)($_POST['gatepass_type_id']??0),
        'department_id'=>$nullableInt($_POST['department_id']??0),
        'requester_type'=>trim((string)($_POST['requester_type']??'employee')),
        'requester_visitor_id'=>$nullableInt($_POST['requester_visitor_id']??0),
        'visit_id'=>$nullableInt($_POST['visit_id']??0),
        'purpose'=>$nullableString($_POST['purpose']??''),
        'is_returnable'=>$returnable,
        'expected_return_at'=>$expected,
        'needs_approval'=>isset($_POST['needs_approval']),
        'items'=>$items,
    ];
    if($payload['gatepass_type_id']<1)$errors[]='Select a gatepass type.';
    if(!in_array($payload['requester_type'],['employee','visitor'],true))$errors[]='Select a valid requester type.';
    if($payload['requester_type']==='visitor' && !$payload['requester_visitor_id'])$errors[]='Select a visitor.';
    if(!$payload['purpose'])$errors[]='Purpose is required.';
    if($returnable && !$expected)$errors[]='Expected return time is required for a returnable gatepass.';
    if(!$errors){
        try{
            $created=Auth::api(App::api(),'POST','/api/v1/gatepasses',$payload);
            $id=apiValue($created,'id');
            flash('success','Gatepass created successfully.');
            redirect($id?'gatepass.php?id='.rawurlencode((string)$id):'gatepasses.php');
        }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to create the gatepass right now.';}
    }
}
App::render('gatepasses/create',compact('errors','types','visitors','departments','currentDepartmentID'));
