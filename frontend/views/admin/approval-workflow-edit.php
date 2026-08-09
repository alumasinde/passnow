<section class="page-header"><div><span class="eyebrow">Administration</span><h1><?= $id?'Edit':'New' ?> approval workflow</h1><p>Define the ordered approval steps returned and enforced by the backend.</p></div><a class="btn btn-secondary" href="<?=e(url('approval-workflows.php'))?>">Back</a></section>
<?php if($errors):?><div class="alert alert-danger"><div><?php foreach($errors as $x):?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="form-grid">
<?php component('field',['name'=>'name','label'=>'Workflow name','value'=>(string)($item['name']??''),'required'=>true]);?>
<?php component('field',['name'=>'description','label'=>'Description','value'=>(string)($item['description']??'')]);?>
<div class="field field-full"><label for="steps_json">Steps JSON</label><textarea id="steps_json" name="steps_json" rows="14" spellcheck="false" required><?=e($stepsJson)?></textarea><small class="field-help">Use the API's step schema. The backend validates ordering, approvers and permissions.</small></div>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('approval-workflows.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Saving..."><span data-button-label>Save workflow</span></button></div>
</form>
