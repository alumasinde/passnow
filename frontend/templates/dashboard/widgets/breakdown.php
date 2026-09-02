<?php
$title=(string)($widget['title']??'Breakdown');
$icon=(string)($widget['icon']??'chart-pie');
$accent=preg_replace('/[^a-z]/','',(string)($widget['accent']??'primary'))?:'primary';
$value=$widget['value']??null;
$data=is_array($widget['data']??null)?$widget['data']:[];
$series=is_array($data['series']??null)?$data['series']:[];
$total=0; foreach($series as $item){$total+=(int)($item['value']??0);}
?>
<article class="dashboard-widget dashboard-widget-content dashboard-widget-breakdown dashboard-accent-<?= e($accent) ?>">
    <header class="dashboard-widget-header">
        <div><span class="dashboard-widget-title"><?= e($title) ?></span><small><?= $value !== null ? e((string)$value).' total' : 'Current distribution' ?></small></div>
        <span class="dashboard-header-icon"><i class="fa-solid fa-<?= e($icon) ?>"></i></span>
    </header>
    <?php if ($series === []): ?>
        <div class="dashboard-empty">No breakdown data available yet.</div>
    <?php else: ?>
        <div class="dashboard-breakdown-list">
            <?php foreach ($series as $item): ?>
                <?php $itemValue=(int)($item['value']??0); $percent=$total>0?round(($itemValue/$total)*100):0; ?>
                <div class="dashboard-breakdown-row">
                    <div class="dashboard-breakdown-top"><span><?= e((string)($item['label']??$item['key']??'Item')) ?></span><strong><?= e((string)$itemValue) ?></strong></div>
                    <div class="dashboard-progress"><span style="width: <?= (int)$percent ?>%"></span></div>
                </div>
            <?php endforeach; ?>
        </div>
    <?php endif; ?>
</article>
