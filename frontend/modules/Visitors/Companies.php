<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=AdminResource::page('/api/v1/visitor-companies',$q);
$columns=[[ 'key'=>'name', 'label'=>'Name' ]];
$actions=[['label'=>'Edit','icon'=>'fa-pen','class'=>'btn-secondary','href'=>fn($row)=>url('visitor-companies-edit.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/visitor-companies',compact('q','r','columns','actions'));
