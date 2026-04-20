const $ = id => document.getElementById(id);
const API = '';  // same origin

let currentFile = null;
let openTabs = [];
let ws = null;

// --- File Explorer ---

async function loadFiles(dir) {
    const param = dir ? `?dir=${encodeURIComponent(dir)}` : '';
    try {
        const res = await fetch(`${API}/api/files${param}`);
        const files = await res.json();
        renderFileTree(files || [], dir);
    } catch (err) {
        console.error('loadFiles error:', err);
    }
}

function renderFileTree(files, parentDir) {
    const tree = $('file-tree');
    tree.innerHTML = '';

    // Back button
    if (parentDir && parentDir !== '.') {
        const back = document.createElement('div');
        back.className = 'file-item dir';
        back.innerHTML = '<span class="file-icon">↩</span><span class="file-label">..</span>';
        const parts = parentDir.split('/');
        parts.pop();
        const parent = parts.join('/') || '.';
        back.addEventListener('click', () => loadFiles(parent));
        tree.appendChild(back);
    }

    // Sort: dirs first, then files
    files.sort((a, b) => {
        if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1;
        return a.name.localeCompare(b.name);
    });

    files.forEach(f => {
        const item = document.createElement('div');
        item.className = `file-item${f.is_dir ? ' dir' : ''}${f.path === currentFile ? ' active' : ''}`;

        const icon = f.is_dir ? '📁' : getFileIcon(f.name);
        item.innerHTML = `<span class="file-icon">${icon}</span><span class="file-label">${f.name}</span>`;

        if (f.is_dir) {
            item.addEventListener('click', () => loadFiles(f.path));
        } else {
            item.addEventListener('click', () => openFile(f.path, f.name));
        }
        tree.appendChild(item);
    });

    if (files.length === 0) {
        tree.innerHTML = '<div class="file-item"><span class="file-label" style="color:var(--text-muted)">Pasta vazia</span></div>';
    }
}

function getFileIcon(name) {
    const ext = name.split('.').pop().toLowerCase();
    const icons = {
        py: '🐍', js: '📜', go: '🔷', ts: '📘', html: '🌐', css: '🎨',
        json: '📋', yaml: '⚙️', yml: '⚙️', md: '📝', txt: '📄',
        sh: '⚡', bash: '⚡', dockerfile: '🐳', mod: '📦', sum: '🔗',
        proto: '📡', sql: '🗄️', csv: '📊', toml: '⚙️'
    };
    return icons[ext] || '📄';
}

// --- Editor ---

async function openFile(path, name) {
    try {
        const res = await fetch(`${API}/api/file?path=${encodeURIComponent(path)}`);
        const content = await res.text();

        currentFile = path;
        $('code-editor').value = content;
        $('current-file-name').textContent = name;
        $('save-indicator').textContent = '';

        // Add tab if not exists
        if (!openTabs.find(t => t.path === path)) {
            openTabs.push({ path, name });
        }
        renderTabs();
        highlightActiveFile();
    } catch (err) {
        termLog('Erro ao abrir: ' + err.message, 'stderr');
    }
}

function renderTabs() {
    const container = $('editor-tabs');
    container.innerHTML = '';

    openTabs.forEach(tab => {
        const el = document.createElement('div');
        el.className = `tab-item${tab.path === currentFile ? ' active' : ''}`;
        el.innerHTML = `<span>${tab.name}</span><span class="tab-close" data-path="${tab.path}">✕</span>`;

        el.addEventListener('click', e => {
            if (e.target.classList.contains('tab-close')) {
                closeTab(e.target.dataset.path);
            } else {
                openFile(tab.path, tab.name);
            }
        });
        container.appendChild(el);
    });
}

function closeTab(path) {
    openTabs = openTabs.filter(t => t.path !== path);
    if (currentFile === path) {
        if (openTabs.length > 0) {
            const last = openTabs[openTabs.length - 1];
            openFile(last.path, last.name);
        } else {
            currentFile = null;
            $('code-editor').value = '';
            $('current-file-name').textContent = 'Nenhum arquivo aberto';
        }
    }
    renderTabs();
}

function highlightActiveFile() {
    document.querySelectorAll('.file-item').forEach(el => {
        el.classList.remove('active');
    });
}

// --- Save ---

async function saveCurrentFile() {
    if (!currentFile) return;
    const content = $('code-editor').value;
    try {
        await fetch(`${API}/api/save`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: currentFile, content })
        });
        $('save-indicator').textContent = '✓ salvo';
        setTimeout(() => $('save-indicator').textContent = '', 2000);
    } catch (err) {
        termLog('Erro ao salvar: ' + err.message, 'stderr');
    }
}

// --- Terminal ---

function termLog(text, type) {
    const output = $('terminal-output');
    const line = document.createElement('div');
    line.className = `term-line term-${type || 'stdout'}`;
    line.textContent = text;
    output.appendChild(line);
    output.scrollTop = output.scrollHeight;
}

function termClear() {
    $('terminal-output').innerHTML = '';
}

function execCommand(command) {
    termLog(`$ ${command}`, 'cmd');

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${location.host}/api/exec`);

    ws.onopen = () => {
        ws.send(JSON.stringify({ command }));
    };

    ws.onmessage = e => {
        try {
            const msg = JSON.parse(e.data);
            if (msg.type === 'stdout') {
                termLog(msg.data, 'stdout');
            } else if (msg.type === 'stderr') {
                termLog(msg.data, 'stderr');
            } else if (msg.type === 'exit') {
                const cls = msg.exit_code === 0 ? 'exit-ok' : 'exit-fail';
                termLog(`Processo finalizado (exit ${msg.exit_code})`, cls);
            } else if (msg.type === 'error') {
                termLog(msg.data, 'stderr');
            }
        } catch {}
    };

    ws.onerror = () => termLog('Erro de conexão WebSocket', 'stderr');
    ws.onclose = () => { ws = null; };
}

function runCurrentFile() {
    if (!currentFile) {
        termLog('Nenhum arquivo aberto para executar.', 'info');
        return;
    }

    saveCurrentFile().then(() => {
        const ext = currentFile.split('.').pop().toLowerCase();
        let cmd;
        switch (ext) {
            case 'py': cmd = `python3 ${currentFile}`; break;
            case 'js': cmd = `node ${currentFile}`; break;
            case 'sh': case 'bash': cmd = `bash ${currentFile}`; break;
            case 'go': cmd = `go run ${currentFile}`; break;
            case 'ts': cmd = `npx ts-node ${currentFile}`; break;
            default: cmd = `cat ${currentFile}`; break;
        }
        execCommand(cmd);
    });
}

// --- Change Directory ---

async function changeDirectory() {
    const path = prompt('Caminho absoluto da pasta:', '/home');
    if (!path) return;
    try {
        const res = await fetch(`${API}/api/setdir`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path })
        });
        const data = await res.json();
        if (res.ok) {
            $('workspace-path').textContent = data.path;
            openTabs = [];
            currentFile = null;
            $('code-editor').value = '';
            $('current-file-name').textContent = 'Nenhum arquivo aberto';
            renderTabs();
            loadFiles('.');
            termLog(`Workspace → ${data.path}`, 'info');
        } else {
            alert(data || 'Pasta não encontrada');
        }
    } catch (err) {
        alert('Erro: ' + err.message);
    }
}

// --- New File ---

function newFile() {
    const name = prompt('Nome do arquivo:', 'script.py');
    if (!name) return;

    // Create empty file
    fetch(`${API}/api/save`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: name, content: '' })
    }).then(() => {
        openFile(name, name);
        loadFiles('.');
        termLog(`Arquivo criado: ${name}`, 'info');
    });
}

// --- Resize Handle ---

let isResizing = false;
const resizeHandle = $('resize-handle');
const editorPanel = $('editor-panel');
const terminalPanel = $('terminal-panel');

resizeHandle.addEventListener('mousedown', e => {
    isResizing = true;
    resizeHandle.classList.add('dragging');
    e.preventDefault();
});

document.addEventListener('mousemove', e => {
    if (!isResizing) return;
    const container = document.querySelector('.editor-area');
    const rect = container.getBoundingClientRect();
    const y = e.clientY - rect.top;
    const total = rect.height;
    const editorH = Math.max(100, Math.min(total - 100, y));
    editorPanel.style.flex = 'none';
    editorPanel.style.height = editorH + 'px';
    terminalPanel.style.height = (total - editorH - 4) + 'px';
});

document.addEventListener('mouseup', () => {
    isResizing = false;
    resizeHandle.classList.remove('dragging');
});

// --- Events ---

$('btn-run').addEventListener('click', runCurrentFile);
$('btn-new-file').addEventListener('click', newFile);
$('btn-refresh-files').addEventListener('click', () => loadFiles('.'));
$('btn-clear-terminal').addEventListener('click', termClear);
$('btn-change-dir').addEventListener('click', changeDirectory);
$('workspace-path').addEventListener('click', changeDirectory);

// Terminal input
$('terminal-input').addEventListener('keydown', e => {
    if (e.key === 'Enter') {
        const cmd = e.target.value.trim();
        if (cmd) {
            execCommand(cmd);
            e.target.value = '';
        }
    }
});

// Keyboard shortcuts
document.addEventListener('keydown', e => {
    // Ctrl+S = save
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        saveCurrentFile();
    }
    // Ctrl+Enter = run
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        runCurrentFile();
    }
});

// Tab key in editor
$('code-editor').addEventListener('keydown', e => {
    if (e.key === 'Tab') {
        e.preventDefault();
        const ta = e.target;
        const start = ta.selectionStart;
        ta.value = ta.value.substring(0, start) + '    ' + ta.value.substring(ta.selectionEnd);
        ta.selectionStart = ta.selectionEnd = start + 4;
    }
});

// Mark unsaved
$('code-editor').addEventListener('input', () => {
    if (currentFile) {
        $('save-indicator').textContent = '● modificado';
        $('save-indicator').style.color = 'var(--yellow)';
    }
});

// --- Init ---

(async function init() {
    // Load current workspace path
    try {
        const res = await fetch(`${API}/api/dir`);
        const data = await res.json();
        $('workspace-path').textContent = data.path;
    } catch {}

    loadFiles('.');
})();
