import { crolab } from "../sdk/src/crolab.js";
const GET = (p) => crolab._api("GET", p);
const POST = (p, b) => crolab._api("POST", p, b);
const PUT = (p, b) => crolab._api("PUT", p, b);
const DEL = (p) => crolab._api("DELETE", p);

// Crolab Admin Panel — Frontend Logic
let currentPlanID = '';


// --- Toast ---
function toast(msg, duration = 3000) {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.classList.remove('hidden');
  clearTimeout(el._t);
  el._t = setTimeout(() => el.classList.add('hidden'), duration);
}

// --- Tab Navigation ---
document.querySelectorAll('.nav-tab').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    showPage(btn.dataset.tab);
  });
});

function showPage(tab) {
  document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
  const section = document.getElementById(tab + '-section');
  if (section) section.classList.remove('hidden');

  if (tab === 'dashboard') loadDashboard();
  if (tab === 'plans') loadPlans();
  if (tab === 'machines') loadMachines();
  if (tab === 'users') loadUsers();
  if (tab === 'providers') loadProviders();
  if (tab === 'logs') loadLogs();
}

// --- Auth ---
async function checkAuth() {
  if (!crolab.token) { showLogin(); return; }
  const { status, data } = await GET('/auth/me');
  if (status !== 200 || data.role !== 'admin') {
    crolab.token = '';
    localStorage.removeItem('admin_token');
    showLogin();
    return;
  }
  document.getElementById('admin-email').textContent = data.email;
  hideLogin();
  showPage('dashboard');
}

function showLogin() {
  document.getElementById('login-section').classList.remove('hidden');
  document.getElementById('admin-nav').style.display = 'none';
  document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
}

function hideLogin() {
  document.getElementById('login-section').classList.add('hidden');
  document.getElementById('admin-nav').style.display = 'flex';
}

document.getElementById('btn-login').addEventListener('click', async () => {
  const email = document.getElementById('login-email').value;
  const password = document.getElementById('login-password').value;
  const { status, data } = await crolab.login(email, password);

  if (status !== 200) {
    document.getElementById('login-error').textContent = data.error || 'Erro no login';
    return;
  }
  if (data.role !== 'admin') {
    document.getElementById('login-error').textContent = 'Acesso restrito a administradores';
    return;
  }
  crolab.token = data.token;
  localStorage.setItem('admin_token', crolab.token);
  document.getElementById('login-error').textContent = '';
  checkAuth();
});

document.getElementById('btn-logout').addEventListener('click', () => {
  crolab.token = '';
  localStorage.removeItem('admin_token');
  showLogin();
});

// --- Dashboard ---
async function loadDashboard() {
  const { data } = await crolab.getDashboard();
  document.getElementById('m-users').textContent = data.users_total ?? '—';
  document.getElementById('m-plans').textContent = data.plans_total ?? '—';
  document.getElementById('m-machines').textContent = data.machines_total ?? '—';
  document.getElementById('m-online').textContent = data.machines_online ?? '—';
}
document.getElementById('btn-refresh-dash').addEventListener('click', loadDashboard);

// --- Plans ---
async function loadPlans() {
  const { data } = await crolab.getPlans();
  const list = document.getElementById('plans-list');
  list.innerHTML = '';

  (data || []).forEach(p => {
    const card = document.createElement('div');
    card.className = 'glass-card plan-card';
    card.innerHTML = `
      <div class="plan-name">${p.name}</div>
      <div class="plan-specs">
        VRAM: ${p.vram || '—'} · Storage: ${p.storage || '—'}<br>
        Max: ${p.max_users} usuários
      </div>
      <div class="plan-price">$${p.price_hr.toFixed(2)}/h <small>· $${p.price_month.toFixed(2)}/mês</small></div>
      <div class="plan-actions">
        <button class="btn-sm" onclick="openPool('${p.id}','${p.name}')">🔗 Pool</button>
        <button class="btn-danger" onclick="deletePlan('${p.id}')">Remover</button>
      </div>
    `;
    list.appendChild(card);
  });
}

document.getElementById('btn-new-plan').addEventListener('click', () => {
  document.getElementById('plan-modal').classList.remove('hidden');
  document.getElementById('plan-modal-title').textContent = 'Novo Plano';
  ['plan-id','plan-name','plan-vram','plan-storage','plan-price-hr','plan-price-month'].forEach(id => {
    document.getElementById(id).value = '';
  });
  document.getElementById('plan-max-users').value = '100';
});

document.getElementById('btn-cancel-plan').addEventListener('click', () => {
  document.getElementById('plan-modal').classList.add('hidden');
});

document.getElementById('btn-save-plan').addEventListener('click', async () => {
  const plan = {
    id: document.getElementById('plan-id').value,
    name: document.getElementById('plan-name').value,
    vram: document.getElementById('plan-vram').value,
    storage: document.getElementById('plan-storage').value,
    price_hr: parseFloat(document.getElementById('plan-price-hr').value) || 0,
    price_month: parseFloat(document.getElementById('plan-price-month').value) || 0,
    max_users: parseInt(document.getElementById('plan-max-users').value) || 100,
  };
  const { status } = await POST('/admin/plans', plan);
  if (status === 201) {
    toast('✅ Plano criado: ' + plan.name);
    document.getElementById('plan-modal').classList.add('hidden');
    loadPlans();
  } else {
    toast('❌ Erro ao criar plano');
  }
});

async function deletePlan(id) {
  if (!confirm('Remover plano ' + id + '?')) return;
  await DEL('/admin/plans/' + id);
  toast('🗑️ Plano removido');
  loadPlans();
}

// --- Pool ---
async function openPool(planID, planName) {
  currentPlanID = planID;
  document.getElementById('pool-plan-name').textContent = planName;
  document.getElementById('pool-section').classList.remove('hidden');
  document.getElementById('pool-form').classList.add('hidden');
  loadPool();
}

async function loadPool() {
  const { data } = await GET('/admin/pool/' + currentPlanID);
  const tbody = document.getElementById('pool-tbody');
  tbody.innerHTML = '';
  (data || []).forEach(e => {
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><strong>${e.priority}</strong></td>
      <td><span class="badge badge-available">${e.provider}</span></td>
      <td>${e.label}</td>
      <td><code>${e.address}</code></td>
      <td><button class="btn-danger" onclick="deletePoolEntry(${e.id})">✕</button></td>
    `;
    tbody.appendChild(tr);
  });
}

document.getElementById('btn-add-pool').addEventListener('click', () => {
  document.getElementById('pool-form').classList.toggle('hidden');
});

document.getElementById('btn-save-pool').addEventListener('click', async () => {
  const entry = {
    priority: parseInt(document.getElementById('pool-priority').value) || 1,
    provider: document.getElementById('pool-provider').value,
    label: document.getElementById('pool-label').value,
    address: document.getElementById('pool-address').value,
    token: document.getElementById('pool-token').value,
  };
  await POST('/admin/pool/' + currentPlanID, entry);
  toast('✅ Entrada adicionada ao pool');
  document.getElementById('pool-form').classList.add('hidden');
  loadPool();
});

async function deletePoolEntry(id) {
  await DEL('/admin/pool/' + currentPlanID + '/' + id);
  toast('🗑️ Entrada removida');
  loadPool();
}

// --- Machines ---
async function loadMachines() {
  const { data } = await crolab.getMachines();
  const tbody = document.getElementById('machines-tbody');
  tbody.innerHTML = '';
  (data || []).forEach(m => {
    const statusClass = m.status === 'available' ? 'badge-available' : 'badge-rented';
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><code>${m.id}</code></td>
      <td>${m.name}</td>
      <td>${m.gpu}</td>
      <td>${m.vram}</td>
      <td class="plan-price">$${m.price_hr.toFixed(2)}</td>
      <td><span class="badge ${statusClass}">${m.status}</span></td>
      <td>${m.provider}</td>
      <td><button class="btn-danger" onclick="deleteMachine('${m.id}')">✕</button></td>
    `;
    tbody.appendChild(tr);
  });
}

document.getElementById('btn-new-machine').addEventListener('click', () => {
  document.getElementById('machine-form').classList.toggle('hidden');
});

document.getElementById('btn-save-machine').addEventListener('click', async () => {
  const m = {
    id: document.getElementById('machine-id').value,
    name: document.getElementById('machine-name').value,
    gpu: document.getElementById('machine-gpu').value,
    vram: document.getElementById('machine-vram').value,
    price_hr: parseFloat(document.getElementById('machine-price').value) || 0,
    address: document.getElementById('machine-address').value,
    provider: document.getElementById('machine-provider').value,
  };
  const { status } = await POST('/admin/machines', m);
  if (status === 201) {
    toast('✅ Máquina adicionada');
    document.getElementById('machine-form').classList.add('hidden');
    loadMachines();
  } else {
    toast('❌ Erro ao adicionar');
  }
});

async function deleteMachine(id) {
  if (!confirm('Remover máquina ' + id + '?')) return;
  await DEL('/admin/machines/' + id);
  toast('🗑️ Máquina removida');
  loadMachines();
}

// --- Users ---
async function loadUsers() {
  const { data } = await GET('/admin/users');
  const tbody = document.getElementById('users-tbody');
  tbody.innerHTML = '';
  (data || []).forEach(u => {
    const roleClass = u.Role === 'admin' ? 'badge-admin' : 'badge-client';
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>${u.ID}</td>
      <td>${u.Email}</td>
      <td><span class="badge ${roleClass}">${u.Role}</span></td>
      <td class="plan-price">$${u.Credits.toFixed(2)}</td>
      <td style="font-size:.75rem;color:var(--text-dim)">${u.CreatedAt}</td>
      <td>
        <button class="btn-sm" onclick="adjustCredits(${u.ID})">💰</button>
        <button class="btn-sm" onclick="toggleRole(${u.ID},'${u.Role}')">🔄</button>
      </td>
    `;
    tbody.appendChild(tr);
  });
}

async function adjustCredits(userID) {
  const amount = prompt('Novo saldo de créditos:');
  if (amount === null) return;
  await PUT('/admin/users/' + userID + '/credits', { credits: parseFloat(amount) });
  toast('✅ Créditos atualizados');
  loadUsers();
}

async function toggleRole(userID, currentRole) {
  const newRole = currentRole === 'admin' ? 'client' : 'admin';
  if (!confirm(`Mudar role para ${newRole}?`)) return;
  await PUT('/admin/users/' + userID + '/role', { role: newRole });
  toast('✅ Role atualizado para ' + newRole);
  loadUsers();
}

// --- Logs ---
async function loadLogs() {
  const { data } = await GET('/admin/logs');
  const tbody = document.getElementById('logs-tbody');
  tbody.innerHTML = '';
  (data || []).forEach(l => {
    let typeClass = 'badge-client';
    if (l.type === 'purchase') typeClass = 'badge-available';
    else if (l.type === 'job') typeClass = 'badge-rented';
    
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><code>${l.id}</code></td>
      <td>${l.email}</td>
      <td><span class="badge ${typeClass}">${l.type}</span></td>
      <td class="plan-price">$${l.amount.toFixed(2)}</td>
      <td>${l.description}</td>
      <td style="font-size:.75rem;color:var(--text-dim)">${new Date(l.created_at).toLocaleString()}</td>
    `;
    tbody.appendChild(tr);
  });
}

document.getElementById('btn-refresh-logs').addEventListener('click', loadLogs);

// Expose to onclick handlers
window.openPool = openPool;
window.deletePlan = deletePlan;
window.deletePoolEntry = deletePoolEntry;
window.deleteMachine = deleteMachine;
window.adjustCredits = adjustCredits;
window.toggleRole = toggleRole;

// --- Init ---
checkAuth();

// --- Providers ---
async function loadProviders() {
  // optionally fetch global configs API here
}

document.getElementById('btn-sync-cloud')?.addEventListener('click', async (e) => {
  const btn = e.currentTarget;
  const originalHtml = btn.innerHTML;
  btn.innerHTML = '<span>⚡ Puxando GPUs (Aguarde)...</span>';
  
  const {status, data} = await crolab.syncCloud();
  btn.innerHTML = originalHtml;
  
  if (status === 200) {
    toast('✅ ' + data.message);
  } else {
    toast('❌ Erro no Sync Cloud');
  }
});
