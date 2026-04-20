*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 10-deploy.md](10-deploy.md) &nbsp; | &nbsp; [Próximo: 12-client-panel.md ➡️](12-client-panel.md)
<hr>

# 11 — Painel Administrativo

## O que é

O Painel Admin é a interface de gerenciamento do Crolab. Acesse via `http://localhost:8844` quando rodando em modo provedor.

## Como acessar

```bash
crolab provider --admin-port :8844 --client-port :8855
```

O primeiro usuário registrado automaticamente recebe role `admin`.

## Funcionalidades

### Dashboard
- Quantidade de usuários, planos, máquinas e GPUs online
- Visão geral do sistema

### Planos (CRUD)
- Criar planos com nome, VRAM, storage, preço/hora e preço/mês
- Cada plano tem um **Pool de Prioridade** associado

### Pool de Prioridade
O core do modelo de negócio. Para cada plano, o admin define uma lista ordenada de provedores:

```
Priority 1: vast-t4-01  (Vast.ai T4, $0.05/h)  ← tenta primeiro (mais barato)
Priority 2: vast-t4-02  (Vast.ai T4, $0.05/h)  ← redundância
Priority 3: runpod-t4   (RunPod T4, $0.12/h)    ← fallback
Priority 4: vps-propria (VPS própria)            ← reserva
```

Quando um job é executado, o orquestrador tenta cada entrada na ordem de prioridade até encontrar uma disponível.

### Máquinas
- Adicionar VPS, servidores dedicados ou GPUs externas
- Monitorar status (available/rented)

### Usuários
- Listar todos os usuários
- Ajustar créditos manualmente
- Alterar role (client → admin ou vice-versa)

## API Endpoints

| Método | Rota | Descrição |
|---|---|---|
| GET | `/admin/dashboard` | Métricas do sistema |
| GET/POST | `/admin/plans` | Listar/criar planos |
| GET/PUT/DELETE | `/admin/plans/:id` | Detalhe/editar/remover |
| GET/POST/DELETE | `/admin/pool/:planID` | Pool de prioridade |
| GET/POST/DELETE | `/admin/machines` | Gerenciar máquinas |
| GET/PUT | `/admin/users` | Gerenciar usuários |
| GET | `/admin/logs` | Log de auditoria |

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 10-deploy.md](10-deploy.md) &nbsp; | &nbsp; [Próximo: 12-client-panel.md ➡️](12-client-panel.md)
