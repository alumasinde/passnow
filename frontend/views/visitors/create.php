<section class="page-header"><div><span class="eyebrow">Visitors</span><h1>New visitor</h1><p>Register a walk-in visitor or pre-register an expected guest.</p></div><a class="btn btn-secondary" data-back href="<?=e(url('visitors.php'))?>"><i class="fa-solid fa-arrow-left"></i> Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
<input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Visitor details</h2><p>Identification requirements adapt to the selected ID type.</p></div>
<div class="form-grid">
<?php component('field',['name'=>'first_name','label'=>'First name','value'=>(string)oldOr('first_name'),'required'=>true,'autocomplete'=>'given-name']); ?>
<?php component('field',['name'=>'last_name','label'=>'Last name','value'=>(string)oldOr('last_name'),'required'=>true,'autocomplete'=>'family-name']); ?>
<div class="field"><label for="id_type_id">ID type <span class="required">*</span></label><select id="id_type_id" name="id_type_id" required><option value="">Select ID type</option><?php foreach($idTypes as $type):?><option value="<?=e((string)($type['id']??''))?>" data-requires-number="<?=!empty($type['requires_number'])?'1':'0'?>" <?=((string)oldOr('id_type_id')===(string)($type['id']??''))?'selected':''?>><?=e((string)($type['name']??''))?></option><?php endforeach;?></select></div>
<div class="field"><label for="id_number">ID number <span class="required" id="id-number-required">*</span></label><input id="id_number" name="id_number" value="<?=e((string)oldOr('id_number'))?>" maxlength="60"><small class="field-help" id="id-number-help">Required when the selected ID type requires a number.</small></div>
<div class="field"><label for="company_id">Company</label><select id="company_id" name="company_id"><option value="">Select existing company</option><?php foreach($companies as $company):$cid=(int)($company['id']??0);?><option value="<?=e((string)$cid)?>" <?=((string)oldOr('company_id')===(string)$cid)?'selected':''?>><?=e((string)($company['name']??''))?></option><?php endforeach;?></select><small class="field-help">If the company is not listed, add it here instead.</small></div><div class="field"><label for="company_name">New company</label><input id="company_name" name="company_name" value="<?=e((string)oldOr('company_name'))?>" placeholder="Type a new visitor company"><small class="field-help">A new company will be created automatically when no existing company is selected.</small></div>
<?php component('field',['name'=>'phone','label'=>'Phone','value'=>(string)oldOr('phone'),'autocomplete'=>'tel']); ?>
<?php component('field',['name'=>'email','label'=>'Email','type'=>'email','value'=>(string)oldOr('email'),'autocomplete'=>'email']); ?>
<?php component('textarea',['name'=>'notes','label'=>'Notes','value'=>(string)oldOr('notes'),'maxlength'=>500,'placeholder'=>'Optional arrival or identification notes']); ?>
<?php if($preRegistrationEnabled): ?><div class="field field-full"><label class="checkbox-row"><input type="checkbox" name="pre_register" value="1" <?=isset($_POST['pre_register'])?'checked':''?>><span><strong>Pre-register this visitor</strong><br><small>The visitor can be scheduled before arriving at the gate.</small></span></label></div><?php endif;?>
</div>
<div class="form-actions"><a class="btn btn-secondary" data-back href="<?=e(url('visitors.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create visitor</span></button></div>
</form>
<script>
(()=>{const select=document.getElementById('id_type_id'),input=document.getElementById('id_number'),required=document.getElementById('id-number-required');const sync=()=>{const opt=select?.selectedOptions?.[0];const needed=opt?.dataset.requiresNumber==='1';input.required=needed;required.hidden=!needed;};select?.addEventListener('change',sync);sync();})();
</script>
