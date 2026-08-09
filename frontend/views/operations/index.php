<section class="page-header">
 <div><span class="eyebrow">Gate operations</span><h1>Gate verification</h1><p>Look up a gatepass by scanning its QR token.</p></div>
</section>
<section class="content-card">
 <div class="card-header"><h2>QR lookup</h2><p>Use your scanner to enter the opaque QR token.</p></div>
 <form method="get" class="operation-search">
  <div class="search-box"><i class="fa-solid fa-qrcode"></i><input name="token" value="<?=e($token)?>" maxlength="128" autocomplete="off" autofocus placeholder="Scan or enter QR token"></div>
  <button class="btn btn-primary" type="submit"><i class="fa-solid fa-magnifying-glass"></i> Verify</button>
 </form>
</section>
<?php if($error): ?><div class="alert alert-danger"><i class="fa-solid fa-circle-exclamation"></i><?=e($error)?></div><?php endif; ?>
<?php if($record): ?>
<?php $status=strtolower((string)($record['status']??'')); $id=(int)($record['id']??0); ?>
<section class="content-card operation-result">
 <div class="card-header detail-header"><div><h2><?=e((string)($record['gatepass_number']??$record['number']??'Gatepass'))?></h2><p>Verify the physical person and items before proceeding.</p></div>
 <span class="status-badge status-<?=e(preg_replace('/[^a-z0-9_-]/i','-',$status))?>"><?=e($record['status']??'Unknown')?></span></div>
 <dl class="detail-list">
  <?php foreach([
   'Person'=>$record['subject_name']??$record['person_name']??'—',
   'Type'=>$record['gatepass_type_name']??$record['type_name']??'—',
   'Direction'=>$record['direction']??'—',
   'Returnable'=>!empty($record['returnable'])?'Yes':'No',
   'Expected return'=>$record['expected_return_at']??'—',
  ] as $label=>$value): ?><div><dt><?=e($label)?></dt><dd><?=e((string)$value)?></dd></div><?php endforeach; ?>
 </dl>
 <div class="form-actions">
  <?php if($id && in_array($status,['approved','ready','issued'],true)): ?>
   <a class="btn btn-primary" href="<?=e(url('gatepass-operation.php?id='.$id.'&operation=check-out'))?>"><i class="fa-solid fa-right-from-bracket"></i> Check out</a>
  <?php elseif($id && in_array($status,['checked_out','out'],true)): ?>
   <a class="btn btn-primary" href="<?=e(url('gatepass-operation.php?id='.$id.'&operation=check-in'))?>"><i class="fa-solid fa-right-to-bracket"></i> Check in</a>
  <?php endif; ?>
  <?php if($id): ?><a class="btn btn-secondary" href="<?=e(url('gatepass.php?id='.$id))?>">Open details</a><?php endif; ?>
 </div>
</section>
<?php endif; ?>
