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
  
  // Restore sidebar when authenticated
  const sidebar = document.getElementById('client-sidebar');
  if (sidebar) sidebar.style.display = 'flex';
  
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
  
  // Hide sidebar entirely when not logged in
  const sidebar = document.getElementById('client-sidebar');
  if (sidebar) sidebar.style.display = 'none';
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
//  COLAB KERNEL ENGINE (Monaco + WebSocket + Multi-Cell)
// =============================================

let kernelWS = null;
const cells = [];
let notebookInitialized = false;

function updateRuntimeStatusUI(state) {
  const dot = document.getElementById('lab-runtime-dot');
  const txt = document.getElementById('lab-runtime-text');
  if(!dot || !txt) return;

  if (state === 'connected') {
      dot.style.background = '#34a853';
      txt.textContent = 'Connected';
  } else if (state === 'busy') {
      dot.style.background = '#fbbc04';
      txt.textContent = 'Busy';
  } else {
      dot.style.background = '#ea4335';
      txt.textContent = 'Disconnected';
  }
}

function connectKernelWS() {
  if (kernelWS && kernelWS.readyState === WebSocket.OPEN) return kernelWS;

  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const url = `${proto}//${window.location.host}/client/lab/exec?token=${encodeURIComponent(TOKEN)}`;

  kernelWS = new WebSocket(url);

  kernelWS.onopen = () => {
    updateRuntimeStatusUI('connected');
  };

  kernelWS.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      const cellId = msg.cell_id;
      if (!cellId) return;
      
      switch (msg.type) {
        case 'stdout':
          appendOutput(cellId, msg.data, '#0f0');
          break;
        case 'stderr':
          appendOutput(cellId, msg.data, '#f55');
          break;
        case 'status':
          appendOutput(cellId, msg.data, '#9aa0a6');
          if(msg.data.includes('Running')) updateRuntimeStatusUI('busy');
          break;
        case 'exit':
          const color = msg.exit_code === 0 ? '#34a853' : '#ea4335';
          appendOutput(cellId, `\n[Kernel] Célula finalizada\n`, color);
          const btn = document.getElementById(cellId === 'primary' ? 'btn-run-cell' : 'btn-run-' + cellId);
          if (btn) btn.classList.remove('playing');
          updateRuntimeStatusUI('connected');
          break;
        case 'error':
          appendOutput(cellId, `[ERRO] ${msg.data}\n`, '#ea4335');
          const btnErr = document.getElementById(cellId === 'primary' ? 'btn-run-cell' : 'btn-run-' + cellId);
          if (btnErr) btnErr.classList.remove('playing');
          updateRuntimeStatusUI('connected');
          break;
      }
    } catch (e) {
      console.warn("Unhalde WS msg:", event.data);
    }
  };

  kernelWS.onclose = () => {
    kernelWS = null;
    updateRuntimeStatusUI('disconnected');
  };

  kernelWS.onerror = () => {
    kernelWS = null;
    updateRuntimeStatusUI('disconnected');
  };

  return kernelWS;
}

function appendOutput(cellId, text, color = '#0f0') {
  const el = document.getElementById(cellId === 'primary' ? 'lab-output' : 'out-' + cellId);
  if (!el) return;

  const span = document.createElement('span');
  span.style.color = color;
  span.textContent = text;
  el.appendChild(span);

  el.scrollTop = el.scrollHeight;
}

function runCell(cellId) {
  if (!TOKEN) { toast('Faça login primeiro'); return; }
  
  const cell = cells.find(c => c.id === cellId);
  if (!cell || !cell.monacoInstance) { toast('Célula inválida'); return; }

  const code = cell.monacoInstance.getValue();
  if (!code.trim()) { toast('Célula vazia'); return; }

  // UI Updates
  const btnShape = document.getElementById(cellId === 'primary' ? 'btn-run-cell' : 'btn-run-' + cellId);
  if (btnShape) btnShape.classList.add('playing');
  
  // Limpa output anterior
  const outputEl = document.getElementById(cellId === 'primary' ? 'lab-output' : 'out-' + cellId);
  if (outputEl) {
      outputEl.innerHTML = '';
      appendOutput(cellId, '[Kernel] Executando célula...\n', '#9aa0a6');
  }

  const ws = connectKernelWS();

  const sendCommand = () => {
    const lang = document.getElementById('lab-runtime-select')?.value || 'python';
    ws.send(JSON.stringify({
      cell_id: cellId,
      code: code,
      language: lang
    }));
  };

  if (ws.readyState === WebSocket.OPEN) {
    sendCommand();
  } else {
    ws.addEventListener('open', sendCommand, { once: true });
  }
}

async function restartKernelUI() {
  if (!TOKEN) return;
  const lang = document.getElementById('lab-runtime-select')?.value || 'python';
  
  toast('Sinalizando o backend para reiniciar container SRE Docker...');
  try {
    const res = await POST('/client/lab/restart-kernel', { language: lang });
    if(res.status === 200) {
      toast('Kernel restartado!');
    } else {
      toast('Erro ao dar restart no Kernel SRE');
    }
  } catch(e) {
    console.error(e);
  }
}

function changeRuntimeLanguage() {
  const lang = document.getElementById('lab-runtime-select')?.value || 'python';
  toast(`Alternando Kernel Sandbox para: ${lang.toUpperCase()}`);
  restartKernelUI();
}

// =============================================
//  LAB UI MENUS (File, Edit, View, Runtime)
// =============================================

function toggleLabMenu(menuId) {
    const target = document.getElementById(menuId);
    const wasHidden = target.classList.contains('hidden');
    
    closeAllLabMenus();
    
    if (wasHidden) {
        target.classList.remove('hidden');
        lucide.createIcons(); // Garante q ícones renderizem
    }
}

function closeAllLabMenus() {
    const menus = document.querySelectorAll('.lab-dropdown');
    menus.forEach(m => m.classList.add('hidden'));
}

// Fechar clicar fora
window.addEventListener('click', function(e) {
    if (!e.target.closest('.lab-menu-item')) {
        closeAllLabMenus();
    }
});

function toggleMonacoMinimap() {
    cells.forEach(c => {
        if (c.monacoInstance) {
            const current = c.monacoInstance.getOption(monaco.editor.EditorOption.minimap).enabled;
            c.monacoInstance.updateOptions({ minimap: { enabled: !current } });
        }
    });
}

function toggleOutput() {
    const primaryOut = document.getElementById('lab-output');
    if(primaryOut) primaryOut.classList.toggle('hidden');
    
    const cellOuts = document.querySelectorAll('.colab-cell-output-wrapper');
    cellOuts.forEach(o => o.classList.toggle('hidden'));
}

function addCell(type = 'code', initialContent = '', forceId = null) {
    const id = forceId || 'cell_' + Date.now() + Math.floor(Math.random() * 1000);
    const cellIdx = cells.length;
    
    const cell = { id, type, content: initialContent, monacoInstance: null };
    cells.push(cell);

    const container = document.getElementById('lab-cells-container');
    const wrapper = document.createElement('div');
    wrapper.className = 'colab-cell-wrapper';
    wrapper.id = 'wrapper-' + id;

    if (type === 'code') {
        const runBtnId = id === 'primary' ? 'btn-run-cell' : 'btn-run-' + id;
        const outId = id === 'primary' ? 'lab-output' : 'out-' + id;
        
        wrapper.innerHTML = `
            <div class="cell-floating-actions">
                <i data-lucide="arrow-up" onclick="moveCellUp('${id}')"></i>
                <i data-lucide="arrow-down" onclick="moveCellDown('${id}')"></i>
                <i data-lucide="trash-2" onclick="deleteCell('${id}')"></i>
                <i data-lucide="more-vertical"></i>
            </div>
            <div class="colab-cell-main">
                <div class="colab-cell-gutter">
                    <div style="font-size: 11px; margin-top:2px; color: #5f6368;">[${cellIdx + 1}]</div>
                    <div class="play-btn-circle" onclick="runCell('${id}')" id="${runBtnId}">
                        <i data-lucide="play" fill="currentColor"></i>
                    </div>
                </div>
                <div class="colab-cell-editor">
                    <div id="monaco-${id}" class="colab-monaco-mount" style="min-height: 100px;"></div>
                </div>
            </div>
            <div class="colab-cell-output-wrapper">
                <div class="colab-output-gutter">
                    <i data-lucide="corner-down-right" class="out-indicator"></i>
                </div>
                <div class="colab-output-content" id="${outId}"></div>
            </div>
        `;
        container.appendChild(wrapper);

        // Instanciação Monaco Editor para Code
        if (typeof monaco !== 'undefined') {
             cell.monacoInstance = monaco.editor.create(document.getElementById('monaco-' + id), {
                value: initialContent,
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
             cell.monacoInstance.addAction({
                id: 'run-cell-' + id,
                label: 'Run Cell',
                keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
                run: () => runCell(id),
             });
             
             // Backwards Compatibility for Tests & Global Ref
             if (id === 'primary') {
                 window.monacoEditor = cell.monacoInstance;
             }
        }
    } else if (type === 'text') {
         wrapper.innerHTML = `
            <div class="cell-floating-actions">
                <i data-lucide="arrow-up" onclick="moveCellUp('${id}')"></i>
                <i data-lucide="arrow-down" onclick="moveCellDown('${id}')"></i>
                <i data-lucide="trash-2" onclick="deleteCell('${id}')"></i>
                <i data-lucide="more-vertical"></i>
            </div>
            <div class="colab-cell-main" style="border: 1px solid var(--border); padding:0; border-radius: 4px;">
                <div class="colab-cell-editor" style="padding: 12px; width: 100%;">
                    <div id="md-view-${id}" class="markdown-body" style="color: #e8eaed;" ondblclick="editMarkdown('${id}')"></div>
                    <div id="monaco-md-${id}" class="colab-monaco-mount hidden" style="min-height: 80px;"></div>
                </div>
            </div>
         `;
         container.appendChild(wrapper);
         
         const mdView = document.getElementById('md-view-' + id);
         mdView.innerHTML = (typeof marked !== 'undefined') ? marked.parse(initialContent || '*Duplo clique para editar texto (Markdown)*') : initialContent;
         
         if (typeof monaco !== 'undefined') {
             cell.monacoInstance = monaco.editor.create(document.getElementById('monaco-md-' + id), {
                value: initialContent,
                language: 'markdown',
                theme: 'vs-dark',
                fontSize: 14,
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                automaticLayout: true,
                padding: { top: 8, bottom: 8 },
                lineNumbers: 'off',
                wordWrap: 'on',
             });
             
             cell.monacoInstance.onDidBlurEditorText(() => {
                 saveMarkdown(id);
             });
         }
    }

    if (window.lucide) lucide.createIcons();
}

function editMarkdown(id) {
    const cell = cells.find(c => c.id === id);
    if(!cell) return;
    document.getElementById('md-view-' + id).classList.add('hidden');
    document.getElementById('monaco-md-' + id).classList.remove('hidden');
    if(cell.monacoInstance) cell.monacoInstance.focus();
}

function saveMarkdown(id) {
    const cell = cells.find(c => c.id === id);
    if(!cell) return;
    const content = cell.monacoInstance.getValue();
    cell.content = content;
    
    document.getElementById('monaco-md-' + id).classList.add('hidden');
    const mdView = document.getElementById('md-view-' + id);
    mdView.classList.remove('hidden');
    
    mdView.innerHTML = content.trim() ? ((typeof marked !== 'undefined') ? marked.parse(content) : content) : '*Duplo clique para editar*';
}

function deleteCell(id) {
    const idx = cells.findIndex(c => c.id === id);
    if (idx > -1) {
        if (cells[idx].monacoInstance) cells[idx].monacoInstance.dispose();
        cells.splice(idx, 1);
        const el = document.getElementById('wrapper-' + id);
        if(el) el.remove();
        updateCellNumbers();
    }
}

function moveCellUp(id) {
    const idx = cells.findIndex(c => c.id === id);
    if (idx > 0) {
        const temp = cells[idx];
        cells[idx] = cells[idx - 1];
        cells[idx - 1] = temp;
        reorderDOM();
    }
}

function moveCellDown(id) {
    const idx = cells.findIndex(c => c.id === id);
    if (idx < cells.length - 1) {
        const temp = cells[idx];
        cells[idx] = cells[idx + 1];
        cells[idx + 1] = temp;
        reorderDOM();
    }
}

function reorderDOM() {
    const container = document.getElementById('lab-cells-container');
    cells.forEach(c => {
        const el = document.getElementById('wrapper-' + c.id);
        if(el) container.appendChild(el);
    });
    updateCellNumbers();
}

function updateCellNumbers() {
    let num = 1;
    cells.forEach(c => {
        if(c.type === 'code') {
            const el = document.querySelector('#wrapper-' + c.id + ' .colab-cell-gutter > div:first-child');
            if(el) el.textContent = '['+num+']';
            num++;
        }
    });
}

function runAllCells() {
    let codeCells = cells.filter(c => c.type === 'code');
    if(codeCells.length === 0) return;
    
    let i = 0;
    function runNext() {
        if (i < codeCells.length) {
            runCell(codeCells[i].id);
            // In a pro engine, we'd wait for exit signal before firing next.
            // Simplified version: wait 200ms and run async
            i++;
            setTimeout(runNext, 400); 
        }
    }
    runNext();
}

function initNotebook() {
    if (notebookInitialized) return;
    const container = document.getElementById('lab-cells-container');
    if (!container) return;
    
    // Wait for Monaco CDN
    if (typeof monaco === 'undefined') {
        setTimeout(initNotebook, 300);
        return;
    }
    
    notebookInitialized = true;
    
    const autosave = localStorage.getItem('crolab_autosave_nb');
    if (autosave && confirm("Encontramos um notebook não salvo previamente. Deseja restaurá-lo?")) {
        loadNotebook(autosave);
    } else {
        addCell('text', '## 👋 Bem-vindo ao Crolab Jupyter Engine!\n\nUse o painel para treinar IAs e executar código. Adicione **Code** ou **Text** pelo Toolbar.', 'cell-intro');
        addCell('code', '# Crolab Jupyter Engine — Escreva seu código aqui\nimport time\n\nfor i in range(5):\n    print(f"[Epoch {i+1}/5] Treinando modelo...")\n    time.sleep(0.3)\n\nprint("✅ Treinamento concluído!")\n', 'primary');
    }
    
    // Bind Toolbar
    document.querySelectorAll('.notebook-toolbar-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const type = e.target.textContent.toLowerCase().includes('text') ? 'text' : 'code';
            addCell(type);
        });
    });
    
    // Bind Add Bottom
    const addBottom = document.querySelector('.add-cell-bottom-bar');
    if(addBottom) {
        addBottom.addEventListener('click', () => addCell('code'));
    }
}



// Hook: initialize Notebook when Lab tab is shown
const origShowPage = showPage;
showPage = function(tab) {
  origShowPage(tab);
  if (tab === 'lab') {
    setTimeout(initNotebook, 200);
  }
};

// Expose
window.subscribePlan = subscribePlan;
window.rentMachine = rentMachine;
window.buyCredits = buyCredits;
window.runCell = runCell;

// Init
if (window.lucide) lucide.createIcons();
checkAuth();

// =============================================
//  NOTEBOOK PERSISTENCE (.ipynb & File API)
// =============================================

let currentFileHandle = null;

function serializeNotebook() {
    const notebook = {
        cells: [],
        metadata: {
            language_info: { name: "python", version: "3" },
            orig_nbformat: 4
        },
        nbformat: 4,
        nbformat_minor: 2
    };

    cells.forEach(c => {
        const source = c.monacoInstance ? c.monacoInstance.getValue() : c.content;
        const cellInfo = {
            cell_type: c.type === 'text' ? 'markdown' : 'code',
            metadata: {},
            source: source.split('\n').map((line, i, arr) => line + (i < arr.length - 1 ? '\n' : '')),
        };
        
        if (c.type === 'code') {
            cellInfo.execution_count = null;
            cellInfo.outputs = []; 
        }
        
        notebook.cells.push(cellInfo);
    });

    return JSON.stringify(notebook, null, 2);
}

function loadNotebook(ipynbContent) {
    try {
        const nb = JSON.parse(ipynbContent);
        if (!nb || !nb.cells) throw new Error("Formato .ipynb inválido");
        
        // Limpa cells atuais
        cells.forEach(c => {
             if (c.monacoInstance) c.monacoInstance.dispose();
             const el = document.getElementById('wrapper-' + c.id);
             if (el) el.remove();
        });
        cells.length = 0; 
        
        nb.cells.forEach(cellObj => {
            const type = cellObj.cell_type === 'markdown' ? 'text' : 'code';
            const content = Array.isArray(cellObj.source) ? cellObj.source.join('') : cellObj.source;
            addCell(type, content);
        });
        
    } catch (e) {
        toast("Erro ao carregar notebook: " + e.message);
    }
}

async function saveNotebookLocal() {
    const jsonStr = serializeNotebook();
    
    if (window.showSaveFilePicker) {
        try {
             if (!currentFileHandle) {
                 currentFileHandle = await window.showSaveFilePicker({
                     suggestedName: 'notebook.ipynb',
                     types: [{ description: 'Jupyter Notebook', accept: { 'application/x-ipynb+json': ['.ipynb'] } }]
                 });
                 document.querySelector('.lab-filename').textContent = currentFileHandle.name + " - CroLab";
             }
             const writable = await currentFileHandle.createWritable();
             await writable.write(jsonStr);
             await writable.close();
             toast("Salvo com sucesso!");
        } catch (e) {
             if (e.name !== 'AbortError') toast("Erro ao salvar: " + e.message);
        }
    } else {
        const blob = new Blob([jsonStr], { type: "application/json" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'notebook.ipynb';
        a.click();
        URL.revokeObjectURL(url);
        toast("Download inicializado (Fallback).");
    }
}

// Global Ctrl+S override e Auto-Save
let autosaveTimer = null;

window.addEventListener('keydown', (e) => {
    // Ctrl+S
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        const labSection = document.getElementById('lab-section');
        if (labSection && !labSection.classList.contains('hidden')) {
             e.preventDefault();
             saveNotebookLocal();
        }
    }
    
    // Auto-save debounce (30s)
    const labSection = document.getElementById('lab-section');
    if (labSection && !labSection.classList.contains('hidden')) {
        clearTimeout(autosaveTimer);
        autosaveTimer = setTimeout(() => {
            localStorage.setItem('crolab_autosave_nb', serializeNotebook());
            console.log("[Auto-save] LocalStorage atualizado");
        }, 30000);
    }
});

// =============================================
//  LOCAL FILE SYSTEM API (Colab Sidebar)
// =============================================
async function mountLocalDrive() {
  try {
    const dHandle = await window.showDirectoryPicker({ mode: 'read' });
    const tree = document.getElementById('lab-file-tree');
    tree.innerHTML = ''; 
    
    const root = document.createElement('div');
    root.className = 'lab-tree-item';
    root.innerHTML = `<i data-lucide="folder-open"></i> <strong>${dHandle.name}</strong>/`;
    tree.appendChild(root);

    for await (const entry of dHandle.values()) {
      const item = document.createElement('div');
      item.className = 'lab-tree-item';
      item.style.cursor = 'pointer';
      
      let icon = 'file';
      if (entry.kind === 'directory') icon = 'folder';
      else if (entry.name.endsWith('.py')) icon = 'file-code-2';
      else if (entry.name.endsWith('.md')) icon = 'book-text';
      else if (entry.name.endsWith('.json') || entry.name.endsWith('.ipynb')) icon = 'file-json-2';

      item.innerHTML = `<i data-lucide="${icon}"></i> ${entry.name}`;
      
      item.addEventListener('dblclick', async () => {
          if (entry.kind === 'file') {
              try {
                  const file = await entry.getFile();
                  const text = await file.text();
                  if (entry.name.endsWith('.ipynb')) {
                      currentFileHandle = entry;
                      loadNotebook(text);
                      const titleEl = document.querySelector('.lab-filename');
                      if (titleEl) titleEl.textContent = entry.name + " - CroLab";
                  } else if (entry.name.endsWith('.py') || entry.name.endsWith('.md')) {
                      const type = entry.name.endsWith('.md') ? 'text' : 'code';
                      addCell(type, text);
                  }
              } catch (e) {
                  toast("Falha ao abrir o arquivo.");
              }
          }
      });

      tree.appendChild(item);
    }
    lucide.createIcons();
    toast('Drive local montado. (Dê duplo clique num .ipynb para abrir)');
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
