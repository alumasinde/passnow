<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Gatepass settings</h1><p>Configure how gatepass numbers are generated for this tenant.</p></div></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
<input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
  <div class="field">
    <label for="number_prefix">Pass number prefix <span class="required">*</span></label>
    <input id="number_prefix" name="number_prefix" type="text" maxlength="32" required value="<?=e((string)($data['number_prefix']??''))?>" placeholder="e.g. GP">
    <small class="field-help">A short tenant-specific prefix placed before every generated gatepass number.</small>
  </div>
  <div class="field field-full">
    <label class="checkbox-row" for="number_use_year">
      <input id="number_use_year" name="number_use_year" type="checkbox" value="1" <?=!empty($data['number_use_year'])?'checked':''?>>
      <span><strong>Include year in pass number</strong><br><small>Example: GP-2026-000001. When disabled, numbering is formatted without the year.</small></span>
    </label>
  </div>
</div>
<div class="form-actions"><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save settings</span></button></div>
</form>