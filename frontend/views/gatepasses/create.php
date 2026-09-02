<section class="page-header">
 <div><span class="eyebrow">Gatepasses</span><h1>New gatepass</h1><p>Create a request using the live rules configured for this tenant.</p></div>
 <a class="btn btn-secondary" data-back href="<?=e(url('gatepasses.php'))?>"><i class="fa-solid fa-arrow-left"></i> Back</a>
</section>
<?php if($errors): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><div><?php foreach($errors as $error): ?><div><?=e($error)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
 <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
 <div class="card-header"><h2>Request details</h2><p>Gatepass behavior is controlled by the selected type.</p></div>
 <div class="form-grid">
  <div class="field"><label>Gatepass type *</label><select name="gatepass_type_id" required data-gatepass-type><option value="">Select gatepass type</option><?php foreach($types as $t): ?><option value="<?=e((string)($t['id']??''))?>" data-direction="<?=e((string)($t['direction']??'out'))?>" data-requires-items="<?=!empty($t['requires_items'])?'1':'0'?>" data-return-policy="<?=e((string)($t['returnability_policy']??'optional'))?>" data-returnable="<?=!empty($t['is_returnable_default'])?'1':'0'?>" <?=((string)($_POST['gatepass_type_id']??'')===(string)($t['id']??''))?'selected':''?>><?=e((string)($t['name']??$t['code']??'Gatepass type'))?></option><?php endforeach;?></select></div>
  <div class="field"><label>Department</label><select name="department_id"><option value="">No department</option><?php foreach($departments as $d): ?><option value="<?=e((string)($d['id']??''))?>"><?=e((string)($d['name']??''))?></option><?php endforeach;?></select></div>
  <div class="field"><label>Requester type *</label><select name="requester_type" required data-requester-type><option value="employee">My employee gatepass</option><option value="visitor">Visitor</option></select><small class="field-help">Employee requests are tied to the signed-in user.</small></div>
  <div class="field" data-visitor-requester hidden><label>Visitor *</label><select name="requester_visitor_id"><option value="">Select visitor</option><?php foreach($visitors as $v): ?><option value="<?=e((string)($v['id']??''))?>"><?=e(trim((string)($v['full_name']??(($v['first_name']??'').' '.($v['last_name']??'')))))?></option><?php endforeach;?></select></div>
  <div class="field field-full"><label>Purpose *</label><textarea name="purpose" required maxlength="255" placeholder="Explain why this gatepass is required"><?=e((string)($_POST['purpose']??''))?></textarea></div>
  <div class="field"><label class="checkbox-field"><input type="checkbox" name="is_returnable" value="1" data-returnable-toggle><span><strong>Returnable</strong><small>Track the item/person return through gate check-in.</small></span></label></div>
  <div class="field" data-return-date><label>Expected return</label><input name="expected_return_at" type="datetime-local"></div>
  <div class="field"><label class="checkbox-field"><input type="checkbox" name="needs_approval" value="1"><span><strong>Request approval</strong><small>The type's mandatory approval policy always takes precedence.</small></span></label></div>
 </div>
 <section class="inline-section" data-items-section>
  <div class="section-heading"><div><h3>Items / assets</h3><p>Add every item covered by this gatepass. Required types enforce at least one item.</p></div><button type="button" class="btn btn-secondary" data-add-gatepass-item><i class="fa-solid fa-plus"></i> Add item</button></div>
  <div data-gatepass-items></div>
 </section>
 <div class="form-actions"><a class="btn btn-secondary" data-back href="<?=e(url('gatepasses.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create gatepass</span></button></div>
</form>
<template id="gatepass-item-template"><div class="content-card item-line" data-gatepass-item><div class="form-grid">
 <div class="field"><label>Name *</label><input name="items[__INDEX__][name]" required></div>
 <div class="field"><label>Quantity *</label><input type="number" min="0.01" step="0.01" name="items[__INDEX__][quantity]" value="1" required></div>
 <div class="field"><label>Direction *</label><select name="items[__INDEX__][direction]"><option value="leaving">Leaving</option><option value="entering">Entering</option><option value="returning">Returning</option></select></div>
 <div class="field"><label>Category</label><input name="items[__INDEX__][category]"></div>
 <div class="field"><label>Unit</label><input name="items[__INDEX__][unit]" placeholder="pcs"></div>
 <div class="field"><label>Asset number</label><input name="items[__INDEX__][asset_number]"></div>
 <div class="field"><label>Serial number</label><input name="items[__INDEX__][serial_number]"></div>
 <div class="field"><label>Condition</label><input name="items[__INDEX__][condition]" placeholder="Good"></div>
 <div class="field field-full"><label>Description</label><input name="items[__INDEX__][description]"></div>
 </div><button type="button" class="btn btn-danger" data-remove-gatepass-item>Remove item</button></div></template>
<script>
(()=>{const form=document.querySelector('form.form-card');const type=form.querySelector('[data-gatepass-type]');const requester=form.querySelector('[data-requester-type]');const visitor=form.querySelector('[data-visitor-requester]');const ret=form.querySelector('[data-returnable-toggle]');const retDate=form.querySelector('[data-return-date]');const items=form.querySelector('[data-gatepass-items]');const section=form.querySelector('[data-items-section]');const tpl=document.getElementById('gatepass-item-template');let i=0;
const sync=()=>{visitor.hidden=requester.value!=='visitor';visitor.querySelector('select').required=requester.value==='visitor';retDate.hidden=!ret.checked;const opt=type.options[type.selectedIndex];const req=opt?.dataset.requiresItems==='1';section.dataset.required=req?'1':'0';};
requester.addEventListener('change',sync);ret.addEventListener('change',sync);type.addEventListener('change',()=>{const o=type.options[type.selectedIndex];if(o){if(o.dataset.returnPolicy==='required')ret.checked=true;if(o.dataset.returnPolicy==='not_allowed')ret.checked=false;}sync();});
form.querySelector('[data-add-gatepass-item]').addEventListener('click',()=>{const node=tpl.content.cloneNode(true);node.firstElementChild.innerHTML=node.firstElementChild.innerHTML.replaceAll('__INDEX__',String(i++));items.append(node);});
items.addEventListener('click',e=>{if(e.target.closest('[data-remove-gatepass-item]'))e.target.closest('[data-gatepass-item]').remove();});sync();})();
</script>