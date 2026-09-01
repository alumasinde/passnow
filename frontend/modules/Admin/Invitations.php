<?php
declare(strict_types=1);require_once __DIR__.'/../app/App.php';Auth::requireLogin();
$q=new ListQuery([]);$r=AdminResource::page('/api/v1/invitations',$q);
$columns=[['key'=>'email','label'=>'Email'],['key'=>'role_name','label'=>'Role'],['key'=>'status','label'=>'Status'],['key'=>'expires_at','label'=>'Expires']];
App::render('admin/invitations',compact('q','r','columns'));
