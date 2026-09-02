<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Visitor settings</h1><p>Configure visitor registration behaviour for this tenant.</p></div></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
<input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<div class="field field-full">
<label class="checkbox-row" for="allow_pre_registration">
<input id="allow_pre_registration" name="allow_pre_registration" type="checkbox" value="1" <?=!empty($data['allow_pre_registration'])?'checked':''?>>
<span><strong>Allow pre-registration</strong><br><small>Allow authorized users to create visitor records before the visitor arrives.</small></span>
</label>
</div>
</div>
<div class="form-actions"><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save settings</span></button></div>
</form>