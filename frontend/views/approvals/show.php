<?php
$number=(string)($approval['gatepass_number']??$approval['number']??('#'.$id));
$status=strtolower((string)($approval['status']??''));
$stepId=(int)($approval['pending_step_id']??$approval['step_id']??0);
?>
<section class="page-header">
 <div><span class="eyebrow">Approval review</span><h1><?=e($number)?></h1><p><?=e((string)($approval['purpose']??'Review gatepass request'))?></p></div>
 <a class="btn btn-secondary" href="<?=e(url('approvals.php'))?>" data-back><i class="fa-solid fa-arrow-left"></i> Back</a>
</section>
<?php if($error): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><?=e($error)?></div><?php endif; ?>

<?php if($approval): ?>
<section class="detail-grid">
 <article class="content-card">
  <div class="card-header detail-header"><div><h2>Request</h2><p>Verify the request before making an approval decision.</p></div>
   <span class="status-badge status-<?=e(preg_replace('/[^a-z0-9_-]/i','-',$status))?>"><?=e($approval['status']??'Unknown')?></span>
  </div>
  <dl class="detail-list">
   <?php foreach([
    'Gatepass number'=>$number,
    'Type'=>$approval['gatepass_type_name']??$approval['type_name']??$approval['type']??'—',
    'Person'=>$approval['subject_name']??$approval['person_name']??'—',
    'Direction'=>$approval['direction']??'—',
    'Purpose'=>$approval['purpose']??'—',
    'Returnable'=>!empty($approval['returnable'])?'Yes':'No',
    'Expected return'=>$approval['expected_return_at']??'—',
    'Created'=>$approval['created_at']??'—',
   ] as $label=>$value): ?><div><dt><?=e($label)?></dt><dd><?=e((string)$value)?></dd></div><?php endforeach; ?>
  </dl>
 </article>

 <article class="content-card">
  <div class="card-header"><h2>Approval chain</h2><p>Steps returned by the backend.</p></div>
  <div class="approval-timeline">
   <?php if(!$steps): ?>
    <?php component('empty-state',['title'=>'No approval steps returned','message'=>'The API did not provide approval history for this request.']); ?>
   <?php else: foreach($steps as $step):
      $stepStatus=strtolower((string)($step['status']??'pending'));
   ?>
    <div class="approval-step">
      <div class="approval-step-marker"><i class="fa-solid <?=in_array($stepStatus,['approved','completed'],true)?'fa-check':'fa-circle'?>"></i></div>
      <div class="approval-step-body">
       <strong><?=e((string)($step['step_name']??$step['name']??'Approval step'))?></strong>
       <span class="status-badge status-<?=e(preg_replace('/[^a-z0-9_-]/i','-',$stepStatus))?>"><?=e($step['status']??'Pending')?></span>
       <?php if(!empty($step['approved_by_name'])): ?><small>By <?=e((string)$step['approved_by_name'])?></small><?php endif; ?>
       <?php if(!empty($step['comment'])): ?><p><?=e((string)$step['comment'])?></p><?php endif; ?>
      </div>
    </div>
   <?php endforeach; endif; ?>
  </div>
 </article>
</section>

<?php if($stepId>0): ?>
<section class="content-card approval-decision">
 <div class="card-header"><h2>Your decision</h2><p>Approval authorization is enforced by the Go API.</p></div>
 <form method="post" action="<?=e(url('approval-decision.php'))?>" data-loading-form>
  <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>">
  <input type="hidden" name="gatepass_id" value="<?=e((string)$id)?>">
  <input type="hidden" name="step_id" value="<?=e((string)$stepId)?>">
  <div class="form-grid">
   <div class="field field-full"><label for="comment">Comment</label><textarea id="comment" name="comment" maxlength="1000" placeholder="Optional approval/rejection comment"></textarea></div>
  </div>
  <div class="form-actions">
   <button class="btn btn-danger" type="submit" name="decision" value="reject" data-loading-label="Rejecting..."><span data-button-label>Reject</span></button>
   <button class="btn btn-primary" type="submit" name="decision" value="approve" data-loading-label="Approving..."><span data-button-label>Approve</span></button>
  </div>
 </form>
</section>
<?php endif; ?>
<?php endif; ?>
