<section class="page-header">
 <div><span class="eyebrow">Administration</span><h1>Settings</h1><p>Configure this tenant without hardcoding operational values.</p></div>
</section>
<div class="settings-grid">
 <?php foreach([
  ['users.php','fa-users','Users','Manage tenant users and memberships.'],
  ['roles.php','fa-shield-halved','Roles & permissions','Configure roles and their permissions.'],
  ['invitations.php','fa-user-plus','Invitations','Invite users into this tenant.'],
  ['approval-workflows.php','fa-list-check','Approval workflows','Configure multi-step approval workflows.'],
  ['gatepass-types.php','fa-ticket','Gatepass types','Configure the types used by gatepasses.'],
  ['gates.php','fa-door-open','Gates','Configure operational gates, entry and exit capabilities.'],
  ['visit-types.php','fa-calendar-check','Visit types','Configure visit categories.'],
  ['id-types.php','fa-id-card','ID types','Manage visitor identification types.'],
  ['visitor-companies.php','fa-building','Visitor companies','Manage visitor company records.'],
  ['departments.php','fa-sitemap','Departments','Manage tenant departments.'],
  ['gatepass-settings.php','fa-sliders','Gatepass settings','Configure numbering, return and operational rules.'],
  ['visitor-settings.php','fa-user-gear','Visitor settings','Configure visitor registration behaviour.'],
  ['theme-settings.php','fa-palette','Theme & branding','Configure tenant identity, logo, colors and appearance.'],
  ['media.php','fa-photo-film','Media library','Upload and manage tenant images and branding files.'],
 ] as [$href,$icon,$title,$description]): ?>
 <a class="settings-card" href="<?=e(url($href))?>">
  <span class="settings-icon"><i class="fa-solid <?=e($icon)?>"></i></span>
  <span><strong><?=e($title)?></strong><small><?=e($description)?></small></span>
  <i class="fa-solid fa-chevron-right settings-arrow"></i>
 </a>
 <?php endforeach; ?>
</div>
