<?php
$number=(string)($gatepass['gatepass_number']??$gatepass['number']??('#'.$id));
$action=$operation==='check-out'?'Check out':'Check in';
$verb=$operation==='check-out'?'leaving the premises':'returning to the premises';
?>
<section class="page-header"><div><span class="eyebrow">Gate operations</span><h1><?=e($action)?></h1><p>Final physical verification before recording the movement.</p></div>
<a class="btn btn-secondary" href="<?=e(url('gatepass.php?id='.$id))?>">Cancel</a></section>
<section class="content-card operation-confirm">
 <div class="operation-icon"><i class="fa-solid <?=$operation==='check-out'?'fa-right-from-bracket':'fa-right-to-bracket'?>"></i></div>
 <h2><?=e($action)?> <?=e($number)?></h2>
 <p>Confirm that the person is physically verified and is <?=e($verb)?>.</p>
 <?php if(!empty($gatepass['returnable']) && $operation==='check-in'): ?><p class="notice"><i class="fa-solid fa-circle-info"></i> Confirm that all returnable items are accounted for before checking in.</p><?php endif; ?>
 <form method="post" data-loading-form>
  <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
  <input type="hidden" name="id" value="<?=e((string)$id)?>">
  <input type="hidden" name="operation" value="<?=e($operation)?>">
  <div class="form-grid"><div class="field"><label>Gate name</label><input name="gate_name" maxlength="120" placeholder="e.g. Main Gate"></div><div class="field field-full"><label>Gate notes</label><textarea name="notes" maxlength="500" placeholder="Optional verification notes"></textarea></div></div>
  <?php if($operation==='check-in'): ?><label class="checkbox-field"><input type="checkbox" name="full_return" value="1" checked><span><strong>Full return verified</strong><small>All outstanding returnable items have been physically verified.</small></span></label><?php endif; ?>
  <div class="form-actions">
   <a class="btn btn-secondary" href="<?=e(url('gatepass.php?id='.$id))?>">Cancel</a>
   <button class="btn btn-primary" type="submit" data-loading-label="<?=e($action.'ing...')?>"><span data-button-label>Confirm <?=e(strtolower($action))?></span></button>
  </div>
 </form>
</section>
