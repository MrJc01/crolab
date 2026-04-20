# /web

Frontends web do Crolab.

## Estrutura

| Pasta | Porta | Descrição |
|---|---|---|
| `admin/` | `:8844` | Painel administrativo (planos, pool, máquinas, usuários) |
| `client/` | `:8855` | Painel do cliente (auth, planos, GPUs, billing) |
| `lab/` | `:8899` | Crolab Lab — editor web com terminal |

## Como servir

```bash
# Modo provedor (admin + client)
crolab provider --admin-port :8844 --client-port :8855

# Modo cloud (frontend único)
crolab cloud-serve --port :8080 --web ./web/client

# Modo lab
crolab lab --port :8899
```
