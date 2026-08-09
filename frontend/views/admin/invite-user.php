<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Invite user</h1><p>Invite a user and assign their tenant role.</p></div><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Back</a></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<?php component('field',['name'=>'first_name','label'=>'First name','required'=>true]);?>
<?php component('field',['name'=>'last_name','label'=>'Last name','required'=>true]);?>
<?php component('field',['name'=>'email','label'=>'Email','type'=>'email','required'=>true]);?>
<?php component('select',['name'=>'role_id','label'=>'Role','options'=>$roles]);?>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Sending..."><span data-button-label>Send invitation</span></button></div>
</form>
