<section class="page-header"><div><span class="eyebrow">Visitors</span><h1>Edit visitor</h1><p>Update the visitor's identification and contact details.</p></div><a class="btn btn-secondary" href="<?=e(url('visitor.php?id='.rawurlencode((string)$id)))?>"><i class="fa-solid fa-arrow-left"></i> Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
<input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Visitor details</h2><p>Changes apply to this visitor in the current tenant database.</p></div>
<div class="form-grid">
<div class="field"><label for="first_name">First name <span class="required">*</span></label><input id="first_name" name="first_name" required autocomplete="given-name" value="<?=e((string)($visitor['first_name']??''))?>"></div>
<div class="field"><label for="last_name">Last name <span class="required">*</span></label><input id="last_name" name="last_name" required autocomplete="family-name" value="<?=e((string)($visitor['last_name']??''))?>"></div>
<div class="field"><label for="id_type_id">ID type <span class="required">*</span></label><select id="id_type_id" name="id_type_id" required><option value="">Select ID type</option><?php foreach($idTypes as $type):$tid=(int)($type['id']??0);?><option value="<?=e((string)$tid)?>" data-requires-number="<?=!empty($type['requires_number'])?'1':'0'?>" <?=((string)($visitor['id_type_id']??'')===(string)$tid)?'selected':''?>><?=e((string)($type['name']??''))?></option><?php endforeach;?></select></div>
<div class="field"><label for="id_number">ID number <span class="required" id="id-number-required">*</span></label><input id="id_number" name="id_number" maxlength="60" value="<?=e((string)($visitor['id_number']??''))?>"><small class="field-help" id="id-number-help">Required when the selected ID type requires a number.</small></div>
<div class="field"><label for="company_id">Company</label><select id="company_id" name="company_id"><option value="">No company</option><?php foreach($companies as $company):$cid=(int)($company['id']??0);?><option value="<?=e((string)$cid)?>" <?=((string)($visitor['company_id']??'')===(string)$cid)?'selected':''?>><?=e((string)($company['name']??''))?></option><?php endforeach;?></select></div>
<div class="field"><label for="phone">Phone</label><input id="phone" name="phone" type="tel" autocomplete="tel" maxlength="30" value="<?=e((string)($visitor['phone']??''))?>"></div>
<div class="field"><label for="email">Email</label><input id="email" name="email" type="email" autocomplete="email" maxlength="255" value="<?=e((string)($visitor['email']??''))?>"></div>
<div class="field field-full"><label for="notes">Notes</label><textarea id="notes" name="notes" maxlength="500" placeholder="Optional arrival or identification notes"><?=e((string)($visitor['notes']??''))?></textarea></div>
</div>
<div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('visitor.php?id='.rawurlencode((string)$id)))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save changes</span></button></div>
</form>
<script>
(()=>{const select=document.getElementById('id_type_id'),input=document.getElementById('id_number'),required=document.getElementById('id-number-required');const sync=()=>{const opt=select?.selectedOptions?.[0];const needed=opt?.dataset.requiresNumber==='1';input.required=needed;required.hidden=!needed;};select?.addEventListener('change',sync);sync();})();
</script>