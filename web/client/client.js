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

  const sidebar = document.querySelector('.colab-sidebar');
  if (sidebar) {
    if (tab === 'lab') {
      sidebar.style.display = 'none';
      if (typeof monacoEditor !== 'undefined' && monacoEditor) setTimeout(() => monacoEditor.layout(), 100);
    } else {
      sidebar.style.display = 'flex';
    }
  }

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
  document.querySelectorAll('.auth-only').forEach(el => el.classList.remove('hidden'));
  
  const hash = window.location.hash.substring(1);
  if (hash && document.getElementById(hash + '-section')) {
    document.querySelectorAll('.nav-tab').forEach(b => b.classList.remove('active'));
    const btn = document.querySelector(`.nav-tab[data-tab="${hash}"]`);
    if(btn) btn.classList.add('active');
    showPage(hash);
  } else {
    showPage('home');
  }
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
  document.querySelectorAll('.auth-only').forEach(el => el.classList.add('hidden'));
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
  toast('Conta criada! 10 créditos de boas-vindas.');
  checkAuth();
});

document.getElementById('btn-logout').addEventListener('click', () => {
  TOKEN = ''; localStorage.removeItem('client_token'); showLanding();
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
    toast(data.message);
    loadSubscription(); loadHome();
  } else {
    toast(data.error || 'Erro ao assinar');
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
    toast(data.message);
    loadGPUs(); loadHome();
  } else {
    toast(data.error || 'Erro');
  }
}

document.getElementById('btn-connect-machine').addEventListener('click', () => {
  document.getElementById('connect-form').classList.toggle('hidden');
});

document.getElementById('btn-do-connect').addEventListener('click', async () => {
  const addr = document.getElementById('conn-address').value;
  const token = document.getElementById('conn-token').value;
  const name = document.getElementById('conn-name').value;
  if (!addr || !name) { toast('Nome e IP obrigatórios'); return; }
  
  const { status, data } = await POST('/client/machines', {
    name, address: addr, token, provider: 'personal', priority: 1
  });
  
  if (status === 201) {
    toast(`Máquina "${name}" conectada.`);
    document.getElementById('connect-form').classList.add('hidden');
  } else {
    toast(`Falha: ${data.error}`);
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
    toast(`$${amount} créditos adicionados`);
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
    toast('Falha ao salvar configuração local');
  }
});

// =============================================
//  COLAB KERNEL ENGINE (Monaco + WebSocket)
// =============================================

let monacoEditor = null;
let kernelWS = null;

function initMonacoEditor() {
  const container = document.getElementById('monaco-container');
  if (!container || monacoEditor) return;

  // Wait for Monaco to load from CDN
  if (typeof monaco === 'undefined') {
    setTimeout(initMonacoEditor, 300);
    return;
  }

  monacoEditor = monaco.editor.create(container, {
    value: '# Crolab Jupyter Engine — Escreva seu código aqui\nimport time\n\nfor i in range(5):\n    print(f"[Epoch {i+1}/5] Treinando modelo...")\n    time.sleep(0.3)\n\nprint("✅ Treinamento concluído!")\n',
    language: 'python',
    theme: 'vs-dark',
    fontSize: 14,
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    automaticLayout: true,
    padding: { top: 12, bottom: 12 },
    lineNumbers: 'on',
    renderLineHighlight: 'all',
    cursorBlinking: 'smooth',
    smoothScrolling: true,
    wordWrap: 'on',
    tabSize: 4,
  });

  // Ctrl+Enter shortcut to run
  monacoEditor.addAction({
    id: 'run-cell',
    label: 'Run Cell',
    keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
    run: () => runClientLabCell(),
  });

  // Watch for resize events to adapt layout
  window.addEventListener('resize', () => {
    if (monacoEditor) monacoEditor.layout();
  });
}

function connectKernelWS() {
  if (kernelWS && kernelWS.readyState === WebSocket.OPEN) return kernelWS;

  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${proto}//${window.location.host}/client/lab/exec?token=${encodeURIComponent(TOKEN)}`;

  kernelWS = new WebSocket(url);

  kernelWS.onopen = () => {
    appendOutput('[Kernel] Conexão WebSocket estabelecida ✓\n', '#0f0');
  };

  kernelWS.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      switch (msg.type) {
        case 'stdout':
          appendOutput(msg.data, '#0f0');
          break;
        case 'stderr':
          appendOutput(msg.data, '#f55');
          break;
        case 'status':
          appendOutput(msg.data + '\n', '#888');
          break;
        case 'exit':
          const color = msg.exit_code === 0 ? '#0f0' : '#f55';
          appendOutput(`\n[Kernel] Processo finalizado (exit code: ${msg.exit_code})\n`, color);
          break;
        case 'error':
          appendOutput(`[ERRO] ${msg.data}\n`, '#f55');
          break;
      }
    } catch (e) {
      appendOutput(event.data + '\n', '#fff');
    }
  };

  kernelWS.onclose = () => {
    appendOutput('[Kernel] Conexão encerrada.\n', '#888');
    kernelWS = null;
  };

  kernelWS.onerror = () => {
    appendOutput('[Kernel] Erro na conexão WebSocket.\n', '#f55');
    kernelWS = null;
  };

  return kernelWS;
}

function appendOutput(text, color = '#0f0') {
  const el = document.getElementById('lab-output');
  if (!el) return;

  const span = document.createElement('span');
  span.style.color = color;
  span.textContent = text;
  el.appendChild(span);

  // Auto-scroll
  el.scrollTop = el.scrollHeight;
}

function runClientLabCell() {
  if (!monacoEditor) { toast('Editor não inicializado'); return; }
  if (!TOKEN) { toast('Faça login primeiro'); return; }

  const code = monacoEditor.getValue();
  if (!code.trim()) { toast('Célula vazia'); return; }

  // UI Updates
  const btnShape = document.getElementById('btn-run-cell');
  if(btnShape) btnShape.classList.add('playing');
  
  // Limpa output anterior
  const outputEl = document.getElementById('lab-output');
  outputEl.innerHTML = '';
  appendOutput('[Kernel] Executando célula...\n', '#9aa0a6'); // gray status

  const ws = connectKernelWS();

  // Escutar quando fechar para parar a animação
  const origClose = ws.onclose;
  ws.onclose = (ev) => {
    if(btnShape) btnShape.classList.remove('playing');
    if(origClose) origClose(ev);
  };

  const sendCommand = () => {
    ws.send(JSON.stringify({
      cell_id: 'cell_' + Date.now(),
      command: `python3 -c ${JSON.stringify(code)}`,
    }));
  };

  if (ws.readyState === WebSocket.OPEN) {
    sendCommand();
  } else {
    ws.addEventListener('open', sendCommand, { once: true });
  }
}

// Hook: initialize Monaco when Lab tab is shown
const origShowPage = showPage;
showPage = function(tab) {
  origShowPage(tab);
  if (tab === 'lab') {
    setTimeout(initMonacoEditor, 200);
  }
};

// Expose
window.subscribePlan = subscribePlan;
window.rentMachine = rentMachine;
window.buyCredits = buyCredits;
window.runClientLabCell = runClientLabCell;

// Init
if (window.lucide) lucide.createIcons();
checkAuth();

// =============================================
//  LOCAL FILE SYSTEM API (Colab Sidebar)
// =============================================
async function mountLocalDrive() {
  try {
    const dHandle = await window.showDirectoryPicker({ mode: 'read' });
    const tree = document.getElementById('lab-file-tree');
    tree.innerHTML = ''; // clear mock
    
    // Add root marker
    const root = document.createElement('div');
    root.className = 'lab-tree-item';
    root.innerHTML = `<i data-lucide="folder-open"></i> <strong>${dHandle.name}</strong>/`;
    tree.appendChild(root);

    for await (const entry of dHandle.values()) {
      const item = document.createElement('div');
      item.className = 'lab-tree-item';
      
      let icon = 'file';
      if (entry.kind === 'directory') icon = 'folder';
      else if (entry.name.endsWith('.py')) icon = 'file-code-2';
      else if (entry.name.endsWith('.md')) icon = 'book-text';
      else if (entry.name.endsWith('.json')) icon = 'file-json-2';

      item.innerHTML = `<i data-lucide="${icon}"></i> ${entry.name}`;
      tree.appendChild(item);
    }
    lucide.createIcons();
    toast('Drive local montado com sucesso.');
  } catch (err) {
    if (err.name !== 'AbortError') {
      toast('Erro ao acessar pasta local: ' + err.message);
    }
  }
}

document.getElementById('btn-mount-local-drive')?.addEventListener('click', mountLocalDrive);

// Sidebar Toggle Logic
document.getElementById('btn-toggle-main-sidebar')?.addEventListener('click', () => {
    const sidebar = document.querySelector('.colab-sidebar');
    if (sidebar) {
        if (sidebar.style.display === 'none') {
            sidebar.style.display = 'flex';
        } else {
            sidebar.style.display = 'none';
        }
        if (typeof monacoEditor !== 'undefined' && monacoEditor) {
            setTimeout(() => monacoEditor.layout(), 100);
        }
    }
});
