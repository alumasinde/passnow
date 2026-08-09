<?php
declare(strict_types=1);
?>
<div class="modal-backdrop" data-modal-backdrop hidden>
  <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
    <div class="modal-header">
      <h2 id="modal-title" data-modal-title>Confirm action</h2>
      <button type="button" class="icon-button" data-modal-close aria-label="Close"><i class="fa-solid fa-xmark"></i></button>
    </div>
    <div class="modal-body" data-modal-body>Are you sure?</div>
    <div class="modal-actions">
      <button type="button" class="btn btn-secondary" data-modal-cancel>Cancel</button>
      <button type="button" class="btn btn-danger" data-modal-confirm>Confirm</button>
    </div>
  </div>
</div>
<div class="page-loading" data-page-loading hidden aria-live="polite">
  <div class="spinner"></div><span>Loading...</span>
</div>
