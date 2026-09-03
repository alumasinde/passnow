<?php
$entry=!array_key_exists('allows_entry',$item)||!empty($item['allows_entry']);
$exit=!array_key_exists('allows_exit',$item)||!empty($item['allows_exit']);
?>
<section class="page-header">
 <div><span class="eyebrow">Settings · Gatepasses</span><h1><?= $id ? 'Edit' : 'Add' ?> gate</h1><p>Every field on this screen maps directly to a backend gate field. No frontend-only gate configuration is stored.</p></div>
 <a class="btn btn-secondary" href="<?=e(url('gates.php'))?>"><i class="fa-solid fa-arrow-left"></i> Back</a>
</section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
 <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
 <div class="form-section"><h2>Gate identity</h2><p class="muted">Use a stable code for integrations and a clear operational name for users.</p>
  <div class="form-grid">
   <div class="field"><label for="gate_name">Name <span class="required">*</span></label><input id="gate_name" name="name" required maxlength="120" value="<?=e((string)oldOr('name',$item['name']??''))?>" placeholder="e.g. Gate A"></div>
   <div class="field"><label for="gate_code">Code <span class="required">*</span></label><input id="gate_code" name="code" required maxlength="30" value="<?=e((string)oldOr('code',$item['code']??''))?>" placeholder="e.g. GATE_A" data-uppercase></div>
   <div class="field"><label for="gate_location">Location / area</label><input id="gate_location" name="location" maxlength="120" value="<?=e((string)oldOr('location',$item['location']??''))?>" placeholder="e.g. Main reception"></div>
   <div class="field field-full"><label for="gate_description">Description</label><textarea id="gate_description" name="description" rows="3" maxlength="255" placeholder="Explain where this gate is and how it is used."><?=e((string)oldOr('description',$item['description']??''))?></textarea></div>
  </div>
 </div>
 <div class="form-section"><h2>Movement capabilities</h2><p class="muted">A gate can allow entry, exit, or both. At least one movement direction is required.</p>
  <div class="form-grid">
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="allows_entry" value="1" <?=$entry?'checked':''?>><span><strong>Allows entry</strong><br><small>This gate can process physical entry movements.</small></span></label></div>
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="allows_exit" value="1" <?=$exit?'checked':''?>><span><strong>Allows exit</strong><br><small>This gate can process physical exit movements.</small></span></label></div>
  </div>
 </div>
 <div class="form-section"><h2>Operational status</h2><div class="form-grid">
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="is_default" value="1" <?=!empty($item['is_default'])?'checked':''?>><span><strong>Default gate</strong><br><small>Only one active tenant gate should be the default. The backend clears the previous default transactionally.</small></span></label></div>
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="active" value="1" <?=(!array_key_exists('active',$item)||!empty($item['active']))?'checked':''?>><span><strong>Gate is active</strong><br><small>Inactive gates remain available in historical records but are excluded from normal operational selection.</small></span></label></div>
 </div></div>
 <div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('gates.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save gate</span></button></div>
</form>
<script>document.querySelector('[data-uppercase]')?.addEventListener('input',e=>e.target.value=e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g,'_'));</script>