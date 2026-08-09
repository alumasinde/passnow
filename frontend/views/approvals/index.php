<section class="page-header">
 <div><span class="eyebrow">Approvals</span><h1>Pending approvals</h1><p>Review gatepasses waiting for your approval.</p></div>
</section>
<?php if($error): ?><div class="alert alert-warning"><i class="fa-solid fa-triangle-exclamation"></i><?=e($error)?></div><?php endif; ?>
<section class="content-card"><div class="toolbar-actions"><button type="button" class="btn btn-secondary" data-export-table="data-table"><i class="fa-solid fa-file-csv"></i> Export</button></div>
 <?php component('list-toolbar',[
   'query'=>$query,
   'filters'=>[['name'=>'status','label'=>'Status','options'=>$statusOptions]],
   'searchPlaceholder'=>'Search gatepass or person...'
 ]); ?>
 <?php component('data-table',[
   'columns'=>$columns,'rows'=>$rows,'rowActions'=>$rowActions,
   'emptyTitle'=>'Nothing needs your approval','emptyMessage'=>'There are no pending approval steps assigned to you.'
 ]); ?>
 <?php component('pagination',['paginator'=>$paginator,'query'=>$query]); ?>
</section>
