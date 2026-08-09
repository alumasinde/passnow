<section class="page-header"><div><span class="eyebrow">Role</span><h1><?=e((string)($role['name']??'Role'))?> permissions</h1><p>Select the permissions assigned to this role.</p></div><a class="btn btn-secondary" href="<?=e(url('roles.php'))?>">Back</a></section>
<?php if($error):?><div class="alert alert-danger"><?=e($error)?></div><?php endif;?>
<form method="post" class="content-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Permissions</h2><p>Permission IDs and names are supplied by the API.</p></div>
<div class="permission-grid"><?php foreach($permissions as $p):$pid=(int)($p['id']??0);$checked=in_array($pid,$selected,true);?><label class="permission-item"><input type="checkbox" name="permissions[]" value="<?=e((string)$pid)?>" <?=$checked?'checked':''?>><span><strong><?=e((string)($p['name']??$p['code']??'Permission'))?></strong><small><?=e((string)($p['description']??''))?></small></span></label><?php endforeach;?></div>
<div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('roles.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save permissions</span></button></div>
</form>
