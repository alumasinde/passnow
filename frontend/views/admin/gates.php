<section class="page-header">
 <div><span class="eyebrow">Settings · Gatepasses</span><h1>Gates</h1><p>Manage the real operational gates used by this tenant. Gate codes and capabilities are controlled by the backend.</p></div>
 <a class="btn btn-primary" href="<?=e(url('gates-edit.php'))?>"><i class="fa-solid fa-plus"></i> Add gate</a>
</section>
<?php if($r['error']):?><div class="alert alert-warning"><?=e($r['error'])?></div><?php endif;?>
<section class="content-card"><?php component('data-table',['columns'=>$columns,'rows'=>$r['rows'],'rowActions'=>$actions,'emptyTitle'=>'No gates configured','emptyMessage'=>'Add Gate A, Gate B, Main Gate or another operational entry point.']);?></section>