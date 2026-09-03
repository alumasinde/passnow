<section class="page-header"><div><span class="eyebrow">Security Operations</span><h1>Gate operations</h1><p>Choose your operating gate, scan or enter a QR token, then process the permitted movement without refreshing the page.</p></div></section>
<?php if($error):?><div class="alert alert-danger"><?=e($error)?></div><?php endif;?>
<section class="detail-grid" data-gate-workspace>
<article class="content-card">
<div class="card-header"><div><h2>Operating gate</h2><p>This gate is attached to every physical movement processed from this workspace.</p></div></div>
<div class="field"><label for="operationGate">Gate</label><select id="operationGate" data-operation-gate><option value="">Select gate</option><?php foreach($gates as $gate):?><option value="<?=e((string)$gate['id'])?>" data-entry="<?=!empty($gate['allows_entry'])?'1':'0'?>" data-exit="<?=!empty($gate['allows_exit'])?'1':'0'?>"><?=e((string)$gate['name'])?><?=!empty($gate['code'])?' · '.e((string)$gate['code']):''?></option><?php endforeach;?></select></div>
<div class="detail-note"><strong>Scanner-ready</strong><p>Use a hardware scanner that types into the token field, or paste a QR token. Camera scanning can be added to this same workflow without changing the movement engine.</p></div>
</article>
<article class="content-card">
<div class="card-header"><div><h2>Find gatepass</h2><p>Scan the QR token or enter it manually.</p></div></div>
<form data-token-form class="form-grid"><div class="field field-full"><label for="gatepassToken">QR token</label><div class="input-action"><input id="gatepassToken" autocomplete="off" placeholder="Scan or enter QR token" data-token-input><button class="btn btn-primary" type="submit">Find</button></div></div></form>
<div data-operation-message></div>
</article>
<article class="content-card" data-record-card hidden>
<div class="card-header"><div><h2 data-record-number>Gatepass</h2><p data-record-status></p></div><span class="status-badge" data-record-badge></span></div>
<dl class="detail-list" data-record-details></dl>
<div class="form-actions"><button class="btn btn-primary" type="button" data-action="checkout">Check out</button><button class="btn btn-primary" type="button" data-action="checkin">Check in</button></div>
</article>
</section>
<script>
(()=>{
const gate=document.querySelector('[data-operation-gate]'),form=document.querySelector('[data-token-form]'),input=document.querySelector('[data-token-input]'),card=document.querySelector('[data-record-card]'),msg=document.querySelector('[data-operation-message]'),number=document.querySelector('[data-record-number]'),status=document.querySelector('[data-record-status]'),badge=document.querySelector('[data-record-badge]'),details=document.querySelector('[data-record-details]');
let record=null;
const key='passnow.operation.gate';
const saved=localStorage.getItem(key);if(saved)gate.value=saved;
gate.addEventListener('change',()=>localStorage.setItem(key,gate.value));
const api=async(url,method='GET',body=null)=>{const r=await fetch(url,{method,headers:{'Content-Type':'application/json','Accept':'application/json'},credentials:'same-origin',body:body?JSON.stringify(body):null});const j=await r.json().catch(()=>({}));if(!r.ok)throw new Error(j.error?.message||j.message||'Operation failed');return j;};
const show=(text,type='info')=>msg.innerHTML='<div class="alert alert-'+(type==='error'?'danger':'success')+'">'+text+'</div>';
form.addEventListener('submit',async e=>{e.preventDefault();const token=input.value.trim();if(!token)return show('Scan or enter a QR token.','error');try{show('Looking up gatepass...');const j=await api('/api/v1/gatepasses/qr/token/'+encodeURIComponent(token));record=j.gatepass||j.data||j;if(!record||!record.id)throw new Error('Gatepass not found');number.textContent=record.pass_number||record.gatepass_number||('#'+record.id);status.textContent=record.purpose||'';badge.textContent=record.status||'Unknown';details.innerHTML='';const rows={'Assigned gate':record.assigned_gate_name||'—','Expected return':record.expected_return_at||'—','Returnable':record.is_returnable?'Yes':'No','Person':record.requester_name||record.subject_name||'—'};Object.entries(rows).forEach(([k,v])=>details.insertAdjacentHTML('beforeend','<div><dt>'+k+'</dt><dd>'+String(v)+'</dd></div>'));card.hidden=false;show('Gatepass found. Verify details before processing.');}catch(err){card.hidden=true;show(err.message,'error');}});
document.querySelectorAll('[data-action]').forEach(btn=>btn.addEventListener('click',async()=>{if(!record)return;const gateID=Number(gate.value);if(!gateID)return show('Select your operating gate first.','error');const op=btn.dataset.action;const selected=gate.options[gate.selectedIndex];if(op==='checkout'&&selected.dataset.exit!=='1')return show('This gate does not allow exit movements.','error');if(op==='checkin'&&selected.dataset.entry!=='1')return show('This gate does not allow entry movements.','error');btn.disabled=true;try{const token=input.value.trim();await api('/api/v1/gatepasses/qr/token/'+encodeURIComponent(token)+'/'+(op==='checkout'?'check-out':'check-in'),'POST',{gate_id:gateID,full_return:op==='checkin'});show('Gatepass '+(op==='checkout'?'checked out':'checked in')+' successfully.');card.hidden=true;input.value='';input.focus();record=null;}catch(err){show(err.message,'error');}finally{btn.disabled=false;}}));
})();
</script>