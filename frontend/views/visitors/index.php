<section class="page-header">
 <div><span class="eyebrow">Visitors</span><h1>Visitors</h1><p>Manage visitors registered within this tenant.</p></div>
 <?php if(Auth::can('visitor.create')): ?><a class="btn btn-primary" href="<?=e(url('visitor-create.php'))?>"><i class="fa-solid fa-user-plus"></i> New visitor</a><?php endif; ?>
</section>
<?php if($error): ?><div class="alert alert-warning"><i class="fa-solid fa-triangle-exclamation"></i><?=e($error)?></div><?php endif; ?>
<section class="content-card"><div class="toolbar-actions"><button type="button" class="btn btn-secondary" data-export-table="data-table"><i class="fa-solid fa-file-csv"></i> Export</button></div>
 <?php component('list-toolbar',['query'=>$query,'filters'=>[
  ['name'=>'status','label'=>'Status','options'=>$statusOptions],
  ['name'=>'blacklisted','label'=>'Blacklist','options'=>$blacklistOptions]
 ],'searchPlaceholder'=>'Search visitor, phone or ID...']); ?>
 <?php component('data-table',['columns'=>$columns,'rows'=>$rows,'rowActions'=>$rowActions,'emptyTitle'=>'No visitors found','emptyMessage'=>'Create a visitor or adjust your search filters.']); ?>
 <?php component('pagination',['paginator'=>$paginator,'query'=>$query]); ?>
</section>
