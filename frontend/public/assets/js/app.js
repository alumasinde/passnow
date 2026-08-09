(() => {
  'use strict';

  const $ = (s, root=document) => root.querySelector(s);
  const $$ = (s, root=document) => [...root.querySelectorAll(s)];

  function setLoading(button, loading, label) {
    if (!button) return;
    button.disabled = loading;
    button.setAttribute('aria-busy', loading ? 'true' : 'false');
    const target = $('[data-button-label]', button);
    if (target) {
      if (loading) {
        button.dataset.originalLabel = target.textContent;
        target.textContent = label || 'Working...';
      } else if (button.dataset.originalLabel) {
        target.textContent = button.dataset.originalLabel;
        delete button.dataset.originalLabel;
      }
    }
  }

  function initForms() {
    $$('form[data-loading-form]').forEach(form => {
      form.addEventListener('submit', () => {
        const button = $('button[type="submit"]', form);
        if (!button) return;
        setLoading(button, true, button.dataset.loadingLabel || 'Working...');
      });
    });
  }

  function initToasts() {
    $$('[data-toast]').forEach(toast => {
      const close = $('[data-toast-close]', toast);
      const remove = () => toast.remove();
      close?.addEventListener('click', remove);
      window.setTimeout(remove, 5000);
    });
  }

  let modalAction = null;
  function openModal({title='Confirm action', body='Are you sure?', confirmText='Confirm', danger=true, onConfirm=null}={}) {
    const backdrop = $('[data-modal-backdrop]');
    if (!backdrop) return;
    $('[data-modal-title]', backdrop).textContent = title;
    $('[data-modal-body]', backdrop).textContent = body;
    const confirm = $('[data-modal-confirm]', backdrop);
    confirm.textContent = confirmText;
    confirm.classList.toggle('btn-danger', danger);
    confirm.classList.toggle('btn-primary', !danger);
    modalAction = onConfirm;
    backdrop.hidden = false;
    document.body.classList.add('modal-open');
    confirm.focus();
  }

  function closeModal() {
    const backdrop = $('[data-modal-backdrop]');
    if (!backdrop) return;
    backdrop.hidden = true;
    document.body.classList.remove('modal-open');
    modalAction = null;
  }

  function initModals() {
    document.addEventListener('click', e => {
      const trigger = e.target.closest('[data-confirm]');
      if (trigger) {
        e.preventDefault();
        openModal({
          title: trigger.dataset.confirmTitle || 'Confirm action',
          body: trigger.dataset.confirm || 'Are you sure you want to continue?',
          confirmText: trigger.dataset.confirmButton || 'Confirm',
          danger: trigger.dataset.confirmDanger !== 'false',
          onConfirm: () => {
            if (trigger.form) {
              trigger.form.submit();
            } else if (trigger.href) {
              window.location.assign(trigger.href);
            }
          }
        });
      }
      if (e.target.closest('[data-modal-close],[data-modal-cancel]')) closeModal();
    });
    $('[data-modal-confirm]')?.addEventListener('click', () => {
      const action = modalAction;
      closeModal();
      if (action) action();
    });
    document.addEventListener('keydown', e => {
      if (e.key === 'Escape') closeModal();
    });
  }

  function initSidebar() {
    const sidebar = $('[data-sidebar]');
    const toggle = $('[data-sidebar-toggle]');
    const backdrop = $('[data-sidebar-backdrop]');
    if (!sidebar || !toggle) return;
    const close = () => {
      sidebar.classList.remove('is-open');
      backdrop?.setAttribute('hidden','');
      document.body.classList.remove('sidebar-open');
    };
    toggle.addEventListener('click', () => {
      const open = sidebar.classList.toggle('is-open');
      if (open) {
        backdrop?.removeAttribute('hidden');
        document.body.classList.add('sidebar-open');
      } else close();
    });
    backdrop?.addEventListener('click', close);
    window.addEventListener('resize', () => { if (window.innerWidth > 900) close(); });
  }

  function initTableFilters() {
    $$('form[data-list-filter]').forEach(form => {
      const select = $('select[name="per_page"]', form);
      select?.addEventListener('change', () => form.submit());
    });
  }


  function csvCell(value) {
    const text = String(value ?? '').replace(/\s+/g,' ').trim();
    return '"' + text.replaceAll('"','""') + '"';
  }
  function initExports() {
    $$('[data-export-table]').forEach(button => {
      button.addEventListener('click', () => {
        const table = document.getElementById(button.dataset.exportTable);
        if (!table) return;
        const rows = $$('tr', table).map(tr => $$('th,td', tr).map(cell => csvCell(cell.innerText)));
        if (!rows.length) return;
        const blob = new Blob([rows.map(r=>r.join(',')).join('\n')], {type:'text/csv;charset=utf-8'});
        const a=document.createElement('a');
        a.href=URL.createObjectURL(blob);
        a.download=(document.title||'passnow').replace(/[^a-z0-9_-]+/gi,'-').toLowerCase()+'.csv';
        document.body.appendChild(a); a.click(); a.remove();
        URL.revokeObjectURL(a.href);
      });
    });
  }

  function init() {
    initForms();
    initToasts();
    initModals();
    initSidebar();
    initTableFilters();
    initExports();
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();

  window.PassNowUI = {openModal, closeModal, setLoading};
})();
