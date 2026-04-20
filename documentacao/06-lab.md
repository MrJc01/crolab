*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 05-cloud-api.md](05-cloud-api.md) &nbsp; | &nbsp; [Próximo: 07-monitor.md ➡️](07-monitor.md)
<hr>

# Crolab Lab — IDE Web (Colab Clone)

## O que é

O Lab é um ambiente notebook web estilo **Google Colab** integrado nativamente no Client Panel. Ele compartilha a mesma porta (`:8855`) e sessão de autenticação do painel do cliente, eliminando a necessidade de um servidor separado.

## Como usar

```bash
# Abre o Lab direto no navegador (Provider Mode)
./crolab lab

# Ou acesse via Client Panel: http://localhost:8855/#lab
```

O navegador abre automaticamente na aba Lab.

## Funcionalidades

### Editor de Código (Monaco Editor)
- Editor Microsoft Monaco (mesmo do VS Code)
- Syntax highlighting para Python, JavaScript, Go, Bash
- Font JetBrains Mono monospace
- Atalho `Ctrl+Enter` para executar célula
- Layout `automaticLayout: true` — se adapta a resizes

### Kernel Stateful (Python)
- Processo Python persistente em background (daemon)
- **Variáveis persistem entre execuções de células** (como Jupyter/Colab)
- Imports são mantidos em memória
- Erros (NameError, TypeError) retornam como stderr sem crashar o kernel
- Streaming em tempo real via WebSocket

### Explorador de Arquivos (Sidebar)
- Botão **Mount Drive** via File System Access API
- Renderiza árvore de arquivos com ícones por tipo (`.py`, `.md`, `.json`)
- Mini-sidebar com ícones de pastas, busca e configurações

### Layout Full-Bleed
- Ao entrar no Lab, a sidebar global do client é automaticamente ocultada
- Botão hamburger (≣) na topbar permite reabrir/fechar a sidebar principal
- Layout ocupa 100% da viewport como Google Colab

### Runtime Status
- Badge de status no canto superior direito
- Indicador visual de conexão WebSocket
- Nome do arquivo (editável) na topbar

## Atalhos

| Atalho | Ação |
|---|---|
| `Ctrl+Enter` | Executar célula |
| Clique ▶ | Executar célula |

## Arquitetura Interna

### Frontend → Backend (WebSocket)

```
Frontend (client.js)                    Backend (server.go)
       │                                      │
       │── WS /client/lab/exec?token=xxx ───→ │
       │                                      │── pythonKernelProxy daemon
       │← JSON {type:"stdout", data:"..."} ──│    (processo persistente)
       │← JSON {type:"stderr", data:"..."} ──│
       │← JSON {type:"exit", exit_code:0} ───│
```

### Protocolo de Mensageria

**Request (Frontend → Backend):**
```json
{
  "cell_id": "cell_1713390000000",
  "code": "print('hello')",
  "language": "python"
}
```

**Response (Backend → Frontend, streaming):**
```json
{"type": "stdout", "data": "hello"}
{"type": "exit", "exit_code": 0}
```

## API Endpoints

| Método | Rota | Descrição |
|---|---|---|
| WS | `/client/lab/exec?token=xxx` | WebSocket para execução stateful |

> **Nota:** O Lab antigo (`internal/lab/server.go`) que expunha uma API REST separada (`/api/files`, `/api/exec`) na porta 19999 foi deprecado. Toda a lógica agora vive integrada no Cloud Server (`internal/cloud/server.go`) na porta 8855.

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 05-cloud-api.md](05-cloud-api.md) &nbsp; | &nbsp; [Próximo: 07-monitor.md ➡️](07-monitor.md)
