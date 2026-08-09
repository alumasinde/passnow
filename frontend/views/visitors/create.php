<section class="page-header"><div><span class="eyebrow">Visitors</span><h1>New visitor</h1><p>Register a visitor for this tenant.</p></div><a class="btn btn-secondary" href="<?=e(url('visitors.php'))?>">Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form>
<input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Visitor details</h2><p>Use the visitor's official identification details.</p></div>
<div class="form-grid">
<?php component('field',['name'=>'first_name','label'=>'First name','value'=>(string)oldOr('first_name'),'required'=>true]); ?>
<?php component('field',['name'=>'last_name','label'=>'Last name','value'=>(string)oldOr('last_name'),'required'=>true]); ?>
<?php component('field',['name'=>'phone','label'=>'Phone','value'=>(string)oldOr('phone'),'required'=>true]); ?>
<?php component('field',['name'=>'email','label'=>'Email','type'=>'email','value'=>(string)oldOr('email')]); ?>
<?php component('select',['name'=>'id_type_id','label'=>'ID type','options'=>$idTypes,'value'=>(string)oldOr('id_type_id'),'required'=>true]); ?>
<?php component('field',['name'=>'id_number','label'=>'ID number','value'=>(string)oldOr('id_number'),'required'=>true]); ?>
<?php component('select',['name'=>'company_id','label'=>'Company','options'=>$companies,'value'=>(string)oldOr('company_id')]); ?>
</div>
<div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('visitors.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create visitor</span></button></div>
</form>
