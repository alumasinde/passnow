<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();
$q=new ListQuery([]);
$r=AdminResource::page('/api/v1/approval-workflows?all=true',$q);
$columns=[
 ['key'=>'name','label'=>'Workflow'],
 ['key'=>'step_count','label'=>'Steps'],
 ['key'=>'active','label'=>'Active'],
];
$actions=[['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('approval-workflow-edit.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/approval-workflows',compact('q','r','columns','actions'));
