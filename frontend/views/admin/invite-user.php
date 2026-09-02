<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Invite user</h1><p>Invite a user and assign their tenant role.</p></div><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Back</a></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<div class="field"><label for="first_name">First name <span class="required">*</span></label><input id="first_name" name="first_name" required autocomplete="given-name" value="<?=e((string)oldOr('first_name'))?>"></div>
<div class="field"><label for="last_name">Last name <span class="required">*</span></label><input id="last_name" name="last_name" required autocomplete="family-name" value="<?=e((string)oldOr('last_name'))?>"></div>
<div class="field"><label for="email">Email <span class="required">*</span></label><input id="email" name="email" type="email" required autocomplete="email" value="<?=e((string)oldOr('email'))?>"></div>
<div class="field"><label for="role_id">Role <span class="required">*</span></label><select id="role_id" name="role_id" required><option value="">Select a role</option><?php foreach($roles as $role):$rid=(int)($role['id']??0);?><option value="<?=e((string)$rid)?>" <?=((int)($_POST['role_id']??0)===$rid)?'selected':''?>><?=e((string)($role['name']??'Role'))?></option><?php endforeach;?></select><?php if(!$roles):?><small class="muted">No roles available. Create a role first.</small><?php endif;?></div>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('users.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Sending..."><span data-button-label>Send invitation</span></button></div>
</form>
