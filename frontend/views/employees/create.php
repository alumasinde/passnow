<section class="page-header"><div><span class="eyebrow">Employees</span><h1>New employee</h1><p>Register an employee for gatepass operations.</p></div><a class="btn btn-secondary" href="<?=e(url('employees.php'))?>">Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Employee details</h2></div><div class="form-grid">
<?php component('field',['name'=>'employee_number','label'=>'Employee number','value'=>(string)oldOr('employee_number'),'required'=>true]);?>
<?php component('field',['name'=>'first_name','label'=>'First name','value'=>(string)oldOr('first_name'),'required'=>true]);?>
<?php component('field',['name'=>'last_name','label'=>'Last name','value'=>(string)oldOr('last_name'),'required'=>true]);?>
<?php component('field',['name'=>'phone','label'=>'Phone','value'=>(string)oldOr('phone')]);?>
<?php component('field',['name'=>'email','label'=>'Email','type'=>'email','value'=>(string)oldOr('email')]);?>
<?php component('select',['name'=>'department_id','label'=>'Department','options'=>$departments,'value'=>(string)oldOr('department_id')]);?>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('employees.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create employee</span></button></div>
</form>
