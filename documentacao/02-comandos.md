*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 01-instalacao.md](01-instalacao.md) &nbsp; | &nbsp; [Próximo: 03-configuracao.md ➡️](03-configuracao.md)
<hr>

# Todos os Comandos CLI

## Visão Geral

```
crolab serve         → Inicia daemon gRPC (recebe jobs)
crolab run           → Envia código para execução remota
crolab monitor       → Dashboard interativo no terminal
crolab lab           → Notebook web (editor + terminal)
crolab config        → CRUD de servidores
crolab auth          → Login/register na Crom Cloud
crolab billing       → Créditos e máquinas
crolab cloud-serve   → Sobe o backend REST API
crolab status        → Mostra estado local
```

---

## `crolab serve`

Inicia o daemon gRPC na máquina atual. Ela se torna receptora de jobs.

```bash
crolab serve --port :4422 --token meu-segredo --slots 3
```

| Flag | Default | Descrição |
|---|---|---|
| `--port` | `:4422` | Porta TCP para escutar |
| `--token` | (vazio) | Token obrigatório para receber jobs |
| `--generate-auth` | false | Gera hash criptográfico automático |
| `--slots` | `2` | Máximo de jobs Docker simultâneos |

Métricas ficam em `http://localhost:9090/metrics`.

---

## `crolab run <diretório>`

Empacota o diretório, envia via gRPC e mostra logs em tempo real.

```bash
crolab run . --image python:3.11-slim --cmd "python train.py" --target meu-gpu
```

| Flag | Default | Descrição |
|---|---|---|
| `--image` | `python:3.11-slim` | Imagem Docker |
| `--cmd` | `ls /workspace` | Comando a executar |
| `--target` | (default) | Nome do servidor. Sem target: seletor interativo |

**Failover**: Se o target falha, tenta o próximo por ordem de prioridade.

---

## `crolab monitor`

Dashboard interativo no terminal (BubbleTea).

| Tecla | Ação |
|---|---|
| ↑ ↓ | Navegar servidores |
| Enter | Definir como default |
| D | Remover servidor |
| A | Adicionar novo (formulário inline) |
| R | Refresh |
| Q | Sair |

Auto-refresh a cada 10 segundos com ping de status (online/offline + latência).

---

## `crolab lab [diretório]`

Abre notebook web no navegador.

```bash
crolab lab .            # Abre pasta atual
crolab lab /home/data   # Abre pasta específica
crolab lab --port :9000 # Porta custom
```

Atalhos no browser: `Ctrl+S` salva, `Ctrl+Enter` executa.

---

## `crolab config`

CRUD de servidores.

```bash
# Adicionar
crolab config add nome ip:porta token --provider vastai --priority 1

# Listar
crolab config ls

# Remover
crolab config rm nome

# Trocar default
crolab config set-default nome
```

---

## `crolab auth`

```bash
crolab auth register email@crom.ai senha123
crolab auth login email@crom.ai senha123
```

Token é salvo automaticamente em `~/.crolab/config.yaml`.

---

## `crolab billing`

```bash
crolab billing status     # Mostra saldo
crolab billing machines   # Lista GPUs disponíveis com preços
```

---

## `crolab cloud-serve`

Sobe o backend REST API (para dev/produção).

```bash
crolab cloud-serve --web ./web
# API: http://localhost:8844
# Frontend: http://localhost:8844
```

---

## `crolab status`

```bash
./crolab status
  Crolab Status
  ─────────────────────────
  OS:       linux/amd64
  Go:       go1.25.0
  Servers:  2 configurados
  Default:  TheTank
  Cloud:    ✓ logado
  GPUs:     1 encontrada(s)
            [0] NVIDIA RTX 4090 (24576 MiB, driver 545.29)
```

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 01-instalacao.md](01-instalacao.md) &nbsp; | &nbsp; [Próximo: 03-configuracao.md ➡️](03-configuracao.md)
