// Crolab Client Panel
const API = window.location.origin;
let TOKEN = localStorage.getItem('client_token') || '';

// --- API ---
async function api(method, path, body) {
  const opts = { method, headers: { 'Content-Type': 'application/json' } };
  if (TOKEN) opts.headers['Authorization'] = TOKEN;
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(API + path, opts);
  return { status: res.status, data: await res.json().catch(() => ({})) };
}
const GET = p => api('GET', p);
const POST = (p, b) => api('POST', p, b);

function toast(msg, dur = 3000) {
  const el = document.getElementById('toast');
  el.textContent = msg; el.classList.remove('hidden');
  clearTimeout(el._t);
  el._t = setTimeout(() => el.classList.add('hidden'), dur);
}

// --- Tabs ---
document.querySelectorAll('.nav-tab').forEach(btn => {
  btn.addEventListener('click', () => {
    if (!TOKEN && btn.dataset.tab !== 'home' && btn.dataset.tab !== 'settings') {
      showPage('auth');
      return;
    }
    document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    showPage(btn.dataset.tab);
  });
});

function showPage(tab) {
  document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
  document.getElementById(tab + '-section')?.classList.remove('hidden');
  if (tab === 'home' && TOKEN) loadHome();
  if (tab === 'settings') loadSettings();
  if (tab === 'plans') { loadClientPlans(); loadSubscription(); }
  if (tab === 'machines') loadGPUs();
  if (tab === 'jobs') loadJobs();
  if (tab === 'billing') loadBilling();
}

// --- Auth ---
async function checkAuth() {
  if (!TOKEN) { showLanding(); return; }
  const { status, data } = await GET('/auth/me');
  if (status !== 200) { TOKEN = ''; localStorage.removeItem('client_token'); showLanding(); return; }
  
  document.getElementById('user-email').textContent = data.email;
  document.getElementById('user-email').classList.remove('hidden');
  document.getElementById('user-credits').textContent = '$' + (data.credits || 0).toFixed(2);
  document.getElementById('user-credits').classList.remove('hidden');
  document.getElementById('btn-logout').classList.remove('hidden');
  document.getElementById('btn-show-auth').classList.add('hidden');
  
  showPage('home');
}

async function tryLocalSSO() {
  try {
    const res = await fetch('/auth/local-token');
    if (res.ok) {
      const data = await res.json();
      if (data.token) { TOKEN = data.token; localStorage.setItem('client_token', TOKEN); }
    }
  } catch (e) {}
}

tryLocalSSO().then(checkAuth);

function showLanding() {
  document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
  document.getElementById('landing-section').classList.remove('hidden');
  
  document.getElementById('user-email').classList.add('hidden');
  document.getElementById('user-credits').classList.add('hidden');
  document.getElementById('btn-logout').classList.add('hidden');
  document.getElementById('btn-show-auth').classList.remove('hidden');
}

document.getElementById('btn-start-landing')?.addEventListener('click', () => {
  if (TOKEN) checkAuth();
  else showPage('auth');
});
document.getElementById('btn-show-auth')?.addEventListener('click', () => {
  showPage('auth');
});

document.getElementById('btn-show-register').addEventListener('click', e => {
  e.preventDefault();
  document.getElementById('auth-login').classList.add('hidden');
  document.getElementById('auth-register').classList.remove('hidden');
});
document.getElementById('btn-show-login').addEventListener('click', e => {
  e.preventDefault();
  document.getElementById('auth-register').classList.add('hidden');
  document.getElementById('auth-login').classList.remove('hidden');
});

document.getElementById('btn-auth-login').addEventListener('click', async () => {
  const email = document.getElementById('auth-email').value;
  const password = document.getElementById('auth-password').value;
  const { status, data } = await POST('/auth/login', { email, password });
  if (status !== 200) { document.getElementById('auth-error').textContent = data.error || 'Erro'; return; }
  TOKEN = data.token; localStorage.setItem('client_token', TOKEN);
  checkAuth();
});

document.getElementById('btn-auth-register').addEventListener('click', async () => {
  const email = document.getElementById('reg-email').value;
  const password = document.getElementById('reg-password').value;
  const { status, data } = await POST('/auth/register', { email, password });
  if (status !== 201) { document.getElementById('auth-error').textContent = data.error || 'Erro'; return; }
  TOKEN = data.token; localStorage.setItem('client_token', TOKEN);
  toast('🎉 Conta criada! 10 créditos de boas-vindas.');
  checkAuth();
});

document.getElementById('btn-logout').addEventListener('click', () => {
  TOKEN = ''; localStorage.removeItem('client_token'); showAuth();
});

// --- Home ---
async function loadHome() {
  const { data: me } = await GET('/auth/me');
  document.getElementById('h-credits').textContent = '$' + (me.credits || 0).toFixed(2);
  document.getElementById('user-credits').textContent = '$' + (me.credits || 0).toFixed(2);

  const { data: sub } = await GET('/client/subscription');
  document.getElementById('h-plan').textContent = sub.plan ? sub.plan.name : 'Nenhum';

  const { data: machines } = await GET('/machines');
  document.getElementById('h-machines').textContent = (machines || []).length;
}

// --- Plans ---
async function loadClientPlans() {
  const { data } = await GET('/client/plans');
  const grid = document.getElementById('plans-grid');
  grid.innerHTML = '';
  (data || []).forEach(p => {
    const card = document.createElement('div');
    card.className = 'glass-card plan-card';
    card.innerHTML = `
      <div class="plan-name">${p.name}</div>
      <div class="plan-specs">
        VRAM: ${p.vram || '—'}<br>
        Storage: ${p.storage || '—'}
      </div>
      <div class="plan-price">$${p.price_hr.toFixed(2)}<span>/hora</span></div>
      ${p.price_month > 0 ? `<div style="font-size:.8rem;color:var(--text-muted);margin-top:.25rem">ou $${p.price_month.toFixed(2)}/mês</div>` : ''}
      <button class="btn-primary" onclick="subscribePlan('${p.id}')">Assinar</button>
    `;
    grid.appendChild(card);
  });
}

async function loadSubscription() {
  const { data } = await GET('/client/subscription');
  const el = document.getElementById('sub-details');
  const btn = document.getElementById('btn-unsubscribe');
  if (data.plan) {
    el.innerHTML = `<strong>${data.plan.name}</strong> — ${data.plan.vram} · $${data.plan.price_hr.toFixed(2)}/h`;
    btn.classList.remove('hidden');
  } else {
    el.textContent = 'Nenhum plano ativo';
    btn.classList.add('hidden');
  }
}

async function subscribePlan(id) {
  const { status, data } = await POST('/client/subscribe', { plan_id: id });
  if (status === 200) {
    toast('✅ ' + data.message);
    loadSubscription(); loadHome();
  } else {
    toast('❌ ' + (data.error || 'Erro ao assinar'));
  }
}

document.getElementById('btn-unsubscribe').addEventListener('click', async () => {
  const { status } = await api('DELETE', '/client/subscription');
  if (status === 200) { toast('Plano cancelado'); loadSubscription(); }
});

// --- Machines ---
async function loadGPUs() {
  const { data } = await GET('/machines');
  const grid = document.getElementById('gpu-grid');
  grid.innerHTML = '';
  (data || []).forEach(m => {
    const card = document.createElement('div');
    card.className = 'glass-card gpu-card';
    const badge = m.status === 'available' ? 'badge-available' : 'badge-rented';
    card.innerHTML = `
      <div class="gpu-name">${m.gpu} <span class="badge ${badge}">${m.status}</span></div>
      <div class="gpu-specs">
        ${m.name}<br>VRAM: ${m.vram} · Provider: ${m.provider || '—'}
      </div>
      <div class="gpu-price">$${m.price_hr.toFixed(2)}/h</div>
      ${m.status === 'available' ? `<button class="btn-primary" style="margin-top:.75rem;width:100%" onclick="rentMachine('${m.id}')">Alugar direto</button>` : ''}
    `;
    grid.appendChild(card);
  });
}

async function rentMachine(id) {
  // Directly rent an available machine bypassing pool
  const { status, data } = await POST('/machines/rent', { machine_id: id });
  if (status === 200) {
    toast('✅ ' + data.message);
    loadGPUs(); loadHome();
  } else {
    toast('❌ ' + (data.error || 'Erro'));
  }
}

document.getElementById('btn-connect-machine').addEventListener('click', () => {
  document.getElementById('connect-form').classList.toggle('hidden');
});

document.getElementById('btn-do-connect').addEventListener('click', async () => {
  const addr = document.getElementById('conn-address').value;
  const token = document.getElementById('conn-token').value;
  const name = document.getElementById('conn-name').value;
  if (!addr || !name) { toast('❌ Nome e IP obrigatórios'); return; }
  
  const { status, data } = await POST('/client/machines', {
    name, address: addr, token, provider: 'personal', priority: 1
  });
  
  if (status === 201) {
    toast(`✅ Máquina "${name}" conectada.`);
    document.getElementById('connect-form').classList.add('hidden');
  } else {
    toast(`❌ Falha: ${data.error}`);
  }
});

// --- Jobs ---
async function loadJobs() {
  const { data } = await GET('/client/jobs');
  const tbody = document.getElementById('jobs-tbody');
  tbody.innerHTML = '';
  (data || []).forEach(j => {
    const card = document.createElement('tr');
    let statusClass = 'badge-available';
    if(j.status === 'queued') statusClass = 'badge-rented';
    if(j.status === 'failed') statusClass = 'badge-danger'; // Requires css, fallback

    card.innerHTML = `
      <td><code>${j.id}</code></td>
      <td>${j.plan_id || '—'}</td>
      <td>${j.machine_used || '—'}</td>
      <td><span class="badge ${statusClass}">${j.status}</span></td>
      <td>${j.duration_s.toFixed(1)}</td>
      <td>$${j.cost.toFixed(4)}</td>
      <td style="font-size:.75rem;color:var(--text-dim)">${new Date(j.created_at).toLocaleString()}</td>
    `;
    tbody.appendChild(card);
  });
}

document.getElementById('btn-refresh-jobs')?.addEventListener('click', loadJobs);

// --- Billing ---
async function loadBilling() {
  const { data: me } = await GET('/auth/me');
  document.getElementById('bill-credits').textContent = '$' + (me.credits || 0).toFixed(2);

  const { data: txs } = await GET('/billing/transactions');
  const tbody = document.getElementById('tx-tbody');
  tbody.innerHTML = '';
  (txs || []).forEach(tx => {
    const tr = document.createElement('tr');
    const color = tx.amount >= 0 ? 'var(--success)' : 'var(--danger)';
    tr.innerHTML = `
      <td style="font-size:.8rem;color:var(--text-dim)">${tx.created_at}</td>
      <td><span class="badge ${tx.type === 'purchase' ? 'badge-available' : 'badge-rented'}">${tx.type}</span></td>
      <td style="color:${color};font-family:var(--mono);font-weight:600">$${tx.amount.toFixed(2)}</td>
      <td>${tx.description}</td>
    `;
    tbody.appendChild(tr);
  });
}

async function buyCredits(amount) {
  const { status } = await POST('/billing/purchase', { amount });
  if (status === 200) {
    toast(`✅ $${amount} créditos adicionados`);
    loadBilling(); loadHome();
  }
}

// --- Settings ---
async function loadSettings() {
  try {
    const res = await fetch(API + '/local-api/config');
    if (res.ok) {
      const data = await res.json();
      document.getElementById('config-target-url').value = data.target || '';
    }
  } catch(e) {}
}

document.getElementById('btn-save-config')?.addEventListener('click', async () => {
  const target = document.getElementById('config-target-url').value;
  if(!target) return;
  const res = await fetch(API + '/local-api/config', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({target})
  });
  if (res.ok) {
    window.location.reload();
  } else {
    toast('❌ Falha ao salvar configuração local');
  }
});

// Expose
window.subscribePlan = subscribePlan;
window.rentMachine = rentMachine;
window.buyCredits = buyCredits;

// Init
checkAuth();
