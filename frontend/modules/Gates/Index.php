<?php
declare(strict_types=1);
require_once __DIR__.'/../../app/App.php';
Auth::requireLogin();

$q=new ListQuery([]);
$r=AdminResource::page('/api/v1/gates?all=true',$q);
$columns=[
 ['key'=>'name','label'=>'Gate'],
 ['key'=>'code','label'=>'Code'],
 ['key'=>'location','label'=>'Location'],
 ['key'=>'allows_entry','label'=>'Entry'],
 ['key'=>'allows_exit','label'=>'Exit'],
 ['key'=>'is_default','label'=>'Default'],
 ['key'=>'active','label'=>'Active'],
];
$actions=[['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('gates-edit.php?id='.rawurlencode((string)($row['id']??'')))]];
App::render('admin/gates',compact('q','r','columns','actions'));