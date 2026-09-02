<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=AdminResource::page('/api/v1/visit-types?all=true',$q);
$columns=[[ 'key'=>'name','label'=>'Name' ],[ 'key'=>'code','label'=>'Code' ],[ 'key'=>'description','label'=>'Description' ],[ 'key'=>'active','label'=>'Active' ]];
$actions=[['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('visit-types-edit.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/visit-types',compact('q','r','columns','actions'));