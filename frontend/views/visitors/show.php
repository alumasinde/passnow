<section class="page-header"><div><span class="eyebrow">Visitor</span><h1><?=e((string)($visitor['full_name']??trim(($visitor['first_name']??'').' '.($visitor['last_name']??''))?:'Visitor'))?></h1><p>Visitor profile and status.</p></div><a class="btn btn-secondary" href="<?=e(url('visitors.php'))?>">Back</a></section>
<?php if($error): ?><div class="alert alert-danger"><?=e($error)?></div><?php endif;?>
<?php if($visitor): $blacklisted=!empty($visitor['blacklisted']); ?>
<section class="content-card"><div class="card-header detail-header"><div><h2>Visitor information</h2><p>Data returned by the tenant API.</p></div><?php if($blacklisted): ?><span class="status-badge status-rejected">Blacklisted</span><?php else: ?><span class="status-badge status-approved">Active</span><?php endif;?></div>
<dl class="detail-list"><?php foreach([
 'First name'=>$visitor['first_name']??'—','Last name'=>$visitor['last_name']??'—','Phone'=>$visitor['phone']??'—',
 'Email'=>$visitor['email']??'—','ID type'=>$visitor['id_type_name']??'—','ID number'=>$visitor['id_number']??'—',
 'Company'=>$visitor['company_name']??'—','Created'=>$visitor['created_at']??'—'
] as $k=>$v):?><div><dt><?=e($k)?></dt><dd><?=e((string)$v)?></dd></div><?php endforeach;?></dl>
<div class="form-actions"><?php if(!$blacklisted): ?><form method="post" action="<?=e(url('visitor-blacklist.php'))?>" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>"><input type="hidden" name="id" value="<?=e((string)$id)?>"><input type="hidden" name="blacklist" value="1"><button class="btn btn-danger" data-loading-label="Updating..."><span data-button-label>Blacklist visitor</span></button></form><?php endif;?></div>
</section>
<?php endif;?>
