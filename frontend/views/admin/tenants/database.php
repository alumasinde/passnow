<?php $title='Tenant database'; ?>
<section class="page-header">
 <div><span class="eyebrow">Platform · Database Operations</span><h1><?=e($tenant['name']??'Tenant')?> database</h1><p>Check isolated tenant database connectivity and safely apply pending tenant migrations.</p></div>
 <div class="inline-actions"><a class="btn btn-secondary" href="<?=e(url('platform/tenant-view?id='.(int)$id))?>">Back to tenant</a><a class="btn btn-secondary" href="<?=e(url('platform/tenant-database?id='.(int)$id))?>">Refresh health</a></div>
</section>
<?php if($error):?><div class="alert alert-danger"><?=e($error)?></div><?php endif?>
<section class="content-card">
 <div class="page-header" style="margin-bottom:1rem"><div><h2>Database health</h2><p class="muted">Live connection check against this tenant's isolated MySQL database.</p></div>
 <span class="badge <?=!empty($health['healthy'])?'badge-success':'badge-danger'?>"><?=!empty($health['healthy'])?'Healthy':'Unhealthy'?></span></div>
 <div class="form-grid">
  <div class="field"><label>Database</label><input readonly value="<?=e((string)($health['database']??'Unavailable'))?>"></div>
  <div class="field"><label>Status</label><input readonly value="<?=e((string)($health['status']??'unknown'))?>"></div>
 </div>
 <p class="muted"><?=e((string)($health['message']??'Health check not available.'))?></p>
</section>
<section class="content-card">
 <h2>Tenant migrations</h2>
 <p class="muted">Runs the repository's <code>migrations/tenant</code> set only against this tenant database. Platform migrations are not touched.</p>
 <?php if(!empty($health['healthy'])):?>
 <form method="post" data-loading-form>
  <input type="hidden" name="_csrf" value="<?=e(Csrf::token())?>"><input type="hidden" name="action" value="migrate">
  <button class="btn btn-primary" data-confirm="Run pending tenant migrations for <?=e($tenant['name']??'this tenant')?>?"><i class="fa-solid fa-database"></i> Run migrations</button>
 </form>
 <?php else:?><div class="alert alert-warning">Fix the tenant database health issue before running migrations.</div><?php endif?>
</section>
