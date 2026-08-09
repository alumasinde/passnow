<section class="page-header"><div><span class="eyebrow">Visits</span><h1>New visit</h1><p>Create a visitor appointment.</p></div><a class="btn btn-secondary" href="<?=e(url('visits.php'))?>">Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Visit details</h2><p>Reference an existing visitor and configure the appointment.</p></div>
<div class="form-grid">
<div class="field"><label for="visitor_id">Visitor ID <span class="required">*</span></label><input id="visitor_id" name="visitor_id" type="number" min="1" required value="<?=e((string)oldOr('visitor_id'))?>"></div>
<?php component('select',['name'=>'visit_type_id','label'=>'Visit type','options'=>$types,'value'=>(string)oldOr('visit_type_id'),'required'=>true]);?>
<?php component('select',['name'=>'department_id','label'=>'Department','options'=>$departments,'value'=>(string)oldOr('department_id')]);?>
<?php component('field',['name'=>'expected_time','label'=>'Expected time','type'=>'datetime-local','value'=>(string)oldOr('expected_time')]);?>
<?php component('textarea',['name'=>'purpose','label'=>'Purpose','value'=>(string)oldOr('purpose'),'required'=>true,'maxlength'=>500,'placeholder'=>'Reason for the visit']);?>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('visits.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create visit</span></button></div>
</form>
