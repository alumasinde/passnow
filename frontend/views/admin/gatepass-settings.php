<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Gatepass settings</h1><p>Values are loaded from the tenant configuration API.</p></div></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<div class="field"><label for="enabled">Enabled</label><select id="enabled" name="enabled"><option value="1" <?=!empty($data['enabled'])?'selected':''?>>Enabled</option><option value="0" <?=isset($data['enabled'])&&!$data['enabled']?'selected':''?>>Disabled</option></select></div>
<div class="field field-full"><label for="configuration">Configuration JSON</label><textarea id="configuration" name="configuration" rows="16" spellcheck="false"><?=e(json_encode($data['configuration']??$data,JSON_PRETTY_PRINT|JSON_UNESCAPED_SLASHES))?></textarea><small class="field-help">The Go API remains the source of truth and validates supported settings.</small></div>
</div><div class="form-actions"><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save settings</span></button></div>
</form>
