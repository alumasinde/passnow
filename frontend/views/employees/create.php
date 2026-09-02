<section class="page-header"><div><span class="eyebrow">Employees</span><h1>New employee</h1><p>Register an employee for gatepass operations.</p></div><a class="btn btn-secondary" href="<?=e(url('employees.php'))?>">Back</a></section>
<?php if($errors): ?><div class="alert alert-danger"><div><?php foreach($errors as $x): ?><div><?=e($x)?></div><?php endforeach;?></div></div><?php endif;?>
<form method="post" class="content-card form-card" data-loading-form><input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
<div class="card-header"><h2>Employee details</h2></div><div class="form-grid">
<div class="field"><label for="employee_number">Employee number <span class="required">*</span></label><input id="employee_number" name="employee_number" type="text" required value="<?=e((string)oldOr('employee_number'))?>"></div>
<div class="field"><label for="first_name">First name <span class="required">*</span></label><input id="first_name" name="first_name" type="text" required autocomplete="given-name" value="<?=e((string)oldOr('first_name'))?>"></div>
<div class="field"><label for="last_name">Last name <span class="required">*</span></label><input id="last_name" name="last_name" type="text" required autocomplete="family-name" value="<?=e((string)oldOr('last_name'))?>"></div>
<div class="field"><label for="phone">Phone</label><input id="phone" name="phone" type="tel" autocomplete="tel" value="<?=e((string)oldOr('phone'))?>"></div>
<div class="field"><label for="email">Email</label><input id="email" name="email" type="email" autocomplete="email" value="<?=e((string)oldOr('email'))?>"></div>
<?php component('select',['name'=>'department_id','label'=>'Department','options'=>$departments,'value'=>(string)oldOr('department_id')]);?>
</div><div class="form-actions"><a class="btn btn-secondary" href="<?=e(url('employees.php'))?>">Cancel</a><button class="btn btn-primary" type="submit" data-loading-label="Creating..."><span data-button-label>Create employee</span></button></div>
</form>
