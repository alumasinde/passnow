<?php
declare(strict_types=1);require_once __DIR__.'/../../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=AdminResource::page('/api/v1/roles',$q);
$columns=[['key'=>'name','label'=>'Role'],['key'=>'description','label'=>'Description'],['key'=>'user_count','label'=>'Users']];
$actions=[['label'=>'Permissions','icon'=>'fa-key','class'=>'btn-secondary','href'=>fn($row)=>url('role-permissions.php?id='.rawurlencode((string)($row['id']??''))) ]];
App::render('admin/roles',compact('q','r','columns','actions'));
