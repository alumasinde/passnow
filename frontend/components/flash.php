<?php
declare(strict_types=1);
$flash = flash_get();
if (!$flash) return;
$type = (string)($flash['type'] ?? 'info');
$message = (string)($flash['message'] ?? '');
?>
<div class="toast toast-<?=e($type)?>" data-toast role="status" aria-live="polite">
    <i class="fa-solid <?=e($type==='success'?'fa-circle-check':($type==='danger'?'fa-circle-xmark':($type==='warning'?'fa-triangle-exclamation':'fa-circle-info')) )?>"></i>
    <span><?=e($message)?></span>
    <button type="button" class="toast-close" data-toast-close aria-label="Dismiss"><i class="fa-solid fa-xmark"></i></button>
</div>
