<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Invite user</h1><p>Invite a user and assign their tenant role.</p></div><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Back</a></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<?php component('field',['name'=>'first_name','label'=>'First name','required'=>true]);?>
<?php component('field',['name'=>'last_name','label'=>'Last name','required'=>true]);?>
<?php component('field',['name'=>'email','label'=>'Email','type'=>'email','required'=>true]);?>
<div class="field"><label for="role_id">Role <span class="required">*</span></label><select id="role_id" name="role_id" required><option value="">Select a role</option><?php foreach($roles as $role):$rid=(int)($role['id']??0);?><option value="<?=e((string)$rid)?>" <?=((int)($_POST['role_id']??0)===$rid)?'selected':''?>><?=e((string)($role['name']??'Role'))?></option><?php endforeach;?></select><?php if(!$roles):?><small class="muted">No roles available. Create a role first.</small><?php endif;?></div>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Sending..."><span data-button-label>Send invitation</span></button></div>
</form>
