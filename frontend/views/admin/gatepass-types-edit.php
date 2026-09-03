<?php
$description=(string)($item['description']??'');
$direction=(string)($item['direction']??'out');
$returnability=(string)($item['returnability_policy']??'optional');
$workflowID=(int)($item['workflow_id']??0);
?>
<section class="page-header">
 <div><span class="eyebrow">Settings · Gatepasses</span><h1><?= $id ? 'Edit' : 'Add' ?> gatepass type</h1><p>Configure how this type behaves at the gate, whether items are required, and whether approval is mandatory.</p></div>
 <a class="btn btn-secondary" href="<?=e(url('gatepass-types.php'))?>"><i class="fa-solid fa-arrow-left"></i> Back</a>
</section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card settings-type-form" data-loading-form>
 <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
 <div class="form-section"><h2>Type identity</h2><p class="muted">Use a stable code for integrations and a clear name for users.</p>
  <div class="form-grid">
   <div class="field"><label for="name">Name <span class="required">*</span></label><input id="name" name="name" required maxlength="120" value="<?=e((string)($item['name']??''))?>" placeholder="e.g. Equipment Removal"></div>
   <div class="field"><label for="code">Code <span class="required">*</span></label><input id="code" name="code" required maxlength="30" value="<?=e((string)($item['code']??''))?>" placeholder="e.g. EQUIP_OUT" data-uppercase><small class="field-help">Short unique identifier. Use letters, numbers and underscores.</small></div>
   <div class="field field-full"><label for="description">Description</label><textarea id="description" name="description" rows="3" maxlength="255" placeholder="Explain when this gatepass type should be used."><?=e($description)?></textarea></div>
  </div>
 </div>
 <div class="form-section"><h2>Gate movement and return rules</h2><p class="muted">These rules control the movement options available for gatepasses of this type.</p>
  <div class="form-grid">
   <div class="field"><label for="direction">Gate direction <span class="required">*</span></label><select id="direction" name="direction" required><option value="out" <?=$direction==='out'?'selected':''?>>Outbound / leaving</option><option value="in" <?=$direction==='in'?'selected':''?>>Inbound / entering</option><option value="both" <?=$direction==='both'?'selected':''?>>Both directions</option></select></div>
   <div class="field"><label for="returnability_policy">Return policy <span class="required">*</span></label><select id="returnability_policy" name="returnability_policy" required data-return-policy><option value="optional" <?=$returnability==='optional'?'selected':''?>>Return is optional</option><option value="required" <?=$returnability==='required'?'selected':''?>>Return is required</option><option value="not_allowed" <?=$returnability==='not_allowed'?'selected':''?>>Return is not allowed</option></select></div>
   <div class="field field-full" data-return-default-row>
    <label class="checkbox-row"><input type="checkbox" name="is_returnable_default" value="1" <?=!empty($item['is_returnable_default'])?'checked':''?> data-return-default><span><strong>Returnable by default</strong><br><small>New gatepasses of this type start as returnable when the policy allows a choice.</small></span></label>
   </div>
  </div>
 </div>
 <div class="form-section"><h2>Gate assignment rules</h2><p class="muted">These fields map directly to the backend gate assignment policy for this gatepass type.</p>
 <div class="form-grid">
  <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="gate_assignment_required" value="1" <?=!empty($item['gate_assignment_required'])?'checked':''?> data-gate-required><span><strong>Assigned gate is required</strong><br><small>A gatepass cannot be created without selecting a gate when this is enabled.</small></span></label></div>
  <div class="field field-full"><label>Allowed gates</label><div class="checkbox-list" data-allowed-gates><?php $selected=array_map('intval',$item['allowed_gate_ids']??[]);foreach($gates as $gate):$gid=(int)($gate['id']??0);?><label class="checkbox-row"><input type="checkbox" name="allowed_gate_ids[]" value="<?=e((string)$gid)?>" <?=in_array($gid,$selected,true)?'checked':''?>><span><strong><?=e((string)($gate['name']??''))?></strong> <small><?=e((string)($gate['code']??''))?><?=!empty($gate['location'])?' · '.e((string)$gate['location']):''?></small></span></label><?php endforeach;if(!$gates):?><small class="field-help">Create an active gate first.</small><?php endif;?></div><small class="field-help">If no allowed gates are selected, any active tenant gate can be assigned. Selecting gates restricts assignment to this list.</small></div>
 </div>
</div>
 <div class="form-section"><h2>Operational requirements</h2>
  <div class="form-grid">
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="requires_items" value="1" <?=!empty($item['requires_items'])?'checked':''?>><span><strong>Items are required</strong><br><small>Users must provide item details before this gatepass can be created.</small></span></label></div>
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="requires_approval" value="1" <?=!empty($item['requires_approval'])?'checked':''?> data-requires-approval><span><strong>Approval is mandatory</strong><br><small>Every gatepass of this type must complete its configured approval workflow.</small></span></label></div>
   <div class="field field-full" data-workflow-row>
    <label for="workflow_id">Approval workflow <span class="required">*</span></label>
    <select id="workflow_id" name="workflow_id" data-workflow-select><option value="">Select workflow</option><?php foreach($workflows as $wf):?><option value="<?=e((string)($wf['id']??''))?>" <?=$workflowID===(int)($wf['id']??0)?'selected':''?>><?=e((string)($wf['name']??''))?><?=(empty($wf['active'])?' (inactive)':'')?></option><?php endforeach;?></select>
    <?php if(!$workflows):?><small class="field-help">No active workflows are available. Create one in Approval workflows first.</small><?php else:?><small class="field-help">The workflow is snapshotted onto each gatepass when it is created.</small><?php endif;?>
   </div>
   <div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="active" value="1" <?=!array_key_exists('active',$item)||!empty($item['active'])?'checked':''?>><span><strong>Type is active</strong><br><small>Inactive types remain in historical records but cannot be selected for new gatepasses.</small></span></label></div>
  </div>
 </div>
 <div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('gatepass-types.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save gatepass type</span></button></div>
</form>
<script>
(()=>{const policy=document.querySelector('[data-return-policy]'),row=document.querySelector('[data-return-default-row]'),check=document.querySelector('[data-return-default]'),approval=document.querySelector('[data-requires-approval]'),workflow=document.querySelector('[data-workflow-row]'),select=document.querySelector('[data-workflow-select]');const syncReturn=()=>{const v=policy.value,locked=v==='required'||v==='not_allowed';row.hidden=locked;if(v==='required'){check.checked=true;check.disabled=true}else if(v==='not_allowed'){check.checked=false;check.disabled=true}else check.disabled=false};const syncApproval=()=>{const on=approval.checked;workflow.hidden=!on;select.required=on;if(!on)select.required=false};policy.addEventListener('change',syncReturn);approval.addEventListener('change',syncApproval);syncReturn();syncApproval();document.querySelector('[data-uppercase]')?.addEventListener('input',e=>e.target.value=e.target.value.toUpperCase().replace(/[^A-Z0-9_]/g,'_'));})();
</script>