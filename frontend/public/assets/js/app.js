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

  function showToast(message, type='success') {
    const old = $('[data-runtime-toast]'); old?.remove();
    const toast = document.createElement('div');
    toast.className = 'toast toast-' + type; toast.dataset.runtimeToast = '1';
    toast.innerHTML = '<span></span><button type="button" class="toast-close" aria-label="Close"><i class="fa-solid fa-xmark"></i></button>';
    toast.querySelector('span').textContent = message || '';
    toast.querySelector('button')?.addEventListener('click', () => toast.remove());
    document.body.appendChild(toast); window.setTimeout(() => toast.remove(), 5000);
  }

  function applyTheme(theme={}) {
    const root = document.documentElement;
    const map = {primary_color:'--color-primary',primary_color_dark:'--color-primary-dark',accent_color:'--color-accent',sidebar_background:'--sidebar-background',sidebar_text:'--sidebar-text',sidebar_active_background:'--sidebar-active-background',sidebar_active_text:'--sidebar-active-text'};
    Object.entries(map).forEach(([key, variable]) => {
      const value = theme[key];
      if (typeof value === 'string' && /^#[0-9a-f]{6}$/i.test(value)) root.style.setProperty(variable, value);
    });
    if (theme.appearance) {
      document.body.classList.remove('theme-light','theme-dark','theme-system');
      document.body.classList.add('theme-' + theme.appearance);
    }
    if (theme.brand_name) $$('[data-tenant-brand-name]').forEach(node => node.textContent = theme.brand_name);
    if (theme.logo_url) $$('[data-tenant-logo]').forEach(img => { img.src = theme.logo_url; });
  }

  function initForms() {
    $$('form[data-loading-form]').forEach(form => {
      form.addEventListener('submit', () => {
        const button = $('button[type="submit"]', form);
        if (!button) return;
        setLoading(button, true, button.dataset.loadingLabel || 'Working...');
      });
    });
    $$('form[data-ajax-form]').forEach(form => {
      form.addEventListener('submit', async event => {
        event.preventDefault();
        const button = $('button[type="submit"]', form);
        setLoading(button, true, button?.dataset.loadingLabel || 'Saving...');
        try {
          const response = await fetch(form.action || window.location.href, {
            method:(form.method || 'POST').toUpperCase(),
            body:new FormData(form),
            credentials:'same-origin',
            headers:{'X-Requested-With':'XMLHttpRequest','Accept':'application/json'}
          });
          const payload = await response.json().catch(() => ({}));
          if (!response.ok || payload.ok === false) throw new Error(payload.message || 'Unable to save changes.');
          const theme = payload.theme || payload.data;
          if (theme && typeof theme === 'object') {
            applyTheme(theme);
            Object.entries(theme).forEach(([key, value]) => {
              const input = form.elements.namedItem(key);
              if (input && typeof value !== 'object') {
                if (input.type === 'checkbox') input.checked = value === true || value === 1 || value === '1' || value === 'true';
                else input.value = value;
              }
              const picker = $('[data-theme-color-picker="' + key + '"]', form);
              if (picker && typeof value === 'string' && /^#[0-9a-f]{6}$/i.test(value)) picker.value = value;
            });
          }
          showToast(payload.message || 'Saved successfully.', 'success');
        } catch (error) {
          showToast(error?.message || 'Unable to save changes.', 'danger');
        } finally { setLoading(button, false); }
      });
    });
  }

  function initUserMenu() {
    const trigger = $('[data-user-menu]'), dropdown = $('[data-user-dropdown]');
    const wrap = $('[data-user-menu-wrap]') || trigger?.closest('.user-menu');
    if (!trigger || !dropdown || !wrap) return;
    const close = () => { dropdown.hidden = true; trigger.setAttribute('aria-expanded','false'); wrap.classList.remove('is-open'); };
    const open = () => { dropdown.hidden = false; trigger.setAttribute('aria-expanded','true'); wrap.classList.add('is-open'); };
    trigger.addEventListener('click', event => { event.stopPropagation(); dropdown.hidden ? open() : close(); });
    dropdown.addEventListener('click', event => event.stopPropagation());
    document.addEventListener('click', close);
    document.addEventListener('keydown', event => { if (event.key === 'Escape') close(); });
  }

  function initBackButtons() {
    $$('[data-back]').forEach(link => link.addEventListener('click', event => {
      const fallback = link.href, referrer = document.referrer;
      let sameOrigin = false; try { sameOrigin = !!referrer && new URL(referrer).origin === window.location.origin; } catch (_) {}
      if (sameOrigin && window.history.length > 1) { event.preventDefault(); window.history.back(); }
      else if (!fallback) event.preventDefault();
    }));
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
        const confirmText = trigger.dataset.confirm || 'Are you sure you want to continue?';
        const backdrop = $('[data-modal-backdrop]');
        e.preventDefault();

        // Some layouts (including Platform Admin) do not render the shared
        // modal component. Fall back to the native confirmation dialog so
        // destructive/important actions still work instead of becoming dead buttons.
        if (!backdrop) {
          if (!window.confirm(confirmText)) return;
          if (trigger.form) {
            if (typeof trigger.form.requestSubmit === 'function') trigger.form.requestSubmit();
            else trigger.form.submit();
          } else if (trigger.href) {
            window.location.assign(trigger.href);
          }
          return;
        }

        openModal({
          title: trigger.dataset.confirmTitle || 'Confirm action',
          body: confirmText,
          confirmText: trigger.dataset.confirmButton || 'Confirm',
          danger: trigger.dataset.confirmDanger !== 'false',
          onConfirm: () => {
            if (trigger.form) {
              if (typeof trigger.form.requestSubmit === 'function') trigger.form.requestSubmit();
              else trigger.form.submit();
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

    const desktop = () => window.innerWidth > 760;
    const sync = () => {
      const collapsed = document.body.classList.contains('sidebar-collapsed');
      const open = sidebar.classList.contains('is-open');
      toggle.setAttribute('aria-expanded', desktop() ? String(!collapsed) : String(open));
      document.body.classList.toggle('sidebar-open', !desktop() && open);
    };

    const closeMobile = () => {
      sidebar.classList.remove('is-open');
      backdrop?.setAttribute('hidden', '');
      document.body.classList.remove('sidebar-open');
      sync();
    };

    try {
      if (localStorage.getItem('passnow.sidebar.collapsed') === '1' && desktop()) {
        document.body.classList.add('sidebar-collapsed');
      }
    } catch (_) {}

    toggle.addEventListener('click', () => {
      if (desktop()) {
        const collapsed = document.body.classList.toggle('sidebar-collapsed');
        try { localStorage.setItem('passnow.sidebar.collapsed', collapsed ? '1' : '0'); } catch (_) {}
        sync();
        return;
      }

      const open = sidebar.classList.toggle('is-open');
      if (open) backdrop?.removeAttribute('hidden'); else backdrop?.setAttribute('hidden', '');
      sync();
    });

    backdrop?.addEventListener('click', closeMobile);
    window.addEventListener('resize', () => {
      if (desktop()) closeMobile();
      sync();
    });
    sync();
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

  function initPermissionUI() {
    $('[data-requires-permission]').forEach(node => {
      const granted = (document.body.dataset.permissions || '').split(',').map(x=>x.trim()).filter(Boolean);
      const required = (node.dataset.requiresPermission || '').split(',').map(x=>x.trim()).filter(Boolean);
      if (!required.length || required.some(p => granted.includes(p))) return;
      if (node.dataset.permissionMode === 'disable') {
        node.setAttribute('aria-disabled','true');
        node.classList.add('is-permission-disabled');
        if ('disabled' in node) node.disabled = true;
      } else node.remove();
    });
  }

  function init() {
    initForms();
    initUserMenu();
    initBackButtons();
    initToasts();
    initModals();
    initSidebar();
    initTableFilters();
    initExports();
    initPermissionUI();
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();

  window.PassNowUI = {openModal, closeModal, setLoading, showToast, applyTheme, initPermissionUI};
})();
