<section class="page-header"><div><span class="eyebrow">Administration</span><h1>Roles & permissions</h1><p>Configure tenant access without embedding role names in the frontend.</p></div></section>
<?php if($r['error']):?><div class="alert alert-warning"><?=e($r['error'])?></div><?php endif;?>
<section class="content-card"><?php component('data-table',['columns'=>$columns,'rows'=>$r['rows'],'rowActions'=>$actions,'emptyTitle'=>'No roles found','emptyMessage'=>'Roles are managed by the tenant administration API.']);?></section>
