<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$id=filter_input(INPUT_GET,'id',FILTER_VALIDATE_INT);
$item=[];$errors=[];$workflows=[];

try{$workflows=apiRows(Auth::api(App::api(),'GET','/api/v1/approval-workflows'));}catch(Throwable){$errors[]='Unable to load approval workflows.';}

if($id){
    try{
        $p=Auth::api(App::api(),'GET','/api/v1/gatepass-types/'.$id);
        $item=apiValue($p,'item',$p['data']??$p);
        if(!is_array($item))$item=[];
    }catch(Throwable $e){$errors[]=$e instanceof ApiException?$e->getMessage():'Unable to load gatepass type.';}
}

if(requestMethod()==='POST'){
    Csrf::requireValid($_POST['_csrf']??null);
    $description=trim((string)($_POST['description']??''));
    $requiresApproval=!empty($_POST['requires_approval']);
    $returnability=trim((string)($_POST['returnability_policy']??'optional'));
    $workflowID=(int)($_POST['workflow_id']??0);

    $payload=[
        'name'=>trim((string)($_POST['name']??'')),
        'code'=>strtoupper(trim((string)($_POST['code']??''))),
        'description'=>$description,
        'direction'=>(string)($_POST['direction']??'out'),
        'returnability_policy'=>$returnability,
        'is_returnable_default'=>!empty($_POST['is_returnable_default']),
        'requires_items'=>!empty($_POST['requires_items']),
        'requires_approval'=>$requiresApproval,
        'active'=>!empty($_POST['active']),
    ];
    if($requiresApproval){
        if($workflowID<1)$errors[]='Select an approval workflow when approval is required.';
        else $payload['workflow_id']=$workflowID;
    }

    if($payload['name']==='')$errors[]='Gatepass type name is required.';
    if($payload['code']==='')$errors[]='Gatepass type code is required.';
    if(!in_array($payload['direction'],['in','out','both'],true))$errors[]='Select a valid gate direction.';
    if(!in_array($returnability,['optional','required','not_allowed'],true))$errors[]='Select a valid return policy.';

    if(!$errors)try{
        if($id){ConfigCrud::update('/api/v1/gatepass-types/'.$id,$payload);flash('success','Gatepass type updated successfully.');}
        else{ConfigCrud::create('/api/v1/gatepass-types',$payload);flash('success','Gatepass type created successfully.');}
        redirect('gatepass-types.php');
    }catch(ApiException $e){$errors[]=$e->getMessage();}catch(Throwable){$errors[]='Unable to save this gatepass type.';}

    $item=array_merge($item,$payload);
    $item['workflow_id']=$workflowID?:null;
}
App::render('admin/gatepass-types-edit',compact('id','item','errors','workflows'));
