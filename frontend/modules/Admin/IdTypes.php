<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=AdminResource::page('/api/v1/id-types',$q);
$columns=[[ 'key'=>'name', 'label'=>'Name' ],[ 'key'=>'code', 'label'=>'Code' ],[ 'key'=>'requires_number', 'label'=>'Requires number' ],[ 'key'=>'active', 'label'=>'Active' ]];
$actions=[['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('id-types-edit.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/id-types',compact('q','r','columns','actions'));
