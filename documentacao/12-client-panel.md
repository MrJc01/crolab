*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 11-admin-panel.md](11-admin-panel.md) &nbsp; | &nbsp; [Próximo: 13-provider-mode.md ➡️](13-provider-mode.md)
<hr>

# 12 — Painel do Cliente

## O que é

O Painel Client é a interface para desenvolvedores que usam o Crolab. Acesse via `http://localhost:8855`.

## Funcionalidades

### Autenticação
- Registro com email + senha (bcrypt)
- Login com token persistente
- 10 créditos de boas-vindas

### Home
- Saldo de créditos
- Plano ativo
- Quantidade de máquinas disponíveis
- Quick Start com comandos CLI

### Planos
- Visualizar planos disponíveis (sem ver detalhes de pool)
- Assinar um plano
- Cancelar assinatura

### Máquinas
- Ver GPUs disponíveis no catálogo
- Alugar GPU diretamente
- Conectar máquina pessoal (IP + token)
- Desconectar

### Billing
- Ver saldo atual
- Comprar créditos ($10, $50, $100)
- Histórico de transações

## API Endpoints

| Método | Rota | Descrição |
|---|---|---|
| POST | `/auth/register` | Criar conta |
| POST | `/auth/login` | Login |
| GET | `/auth/me` | Perfil do user |
| GET | `/client/plans` | Planos disponíveis |
| POST | `/client/subscribe` | Assinar plano |
| GET/DELETE | `/client/subscription` | Ver/cancelar assinatura |
| GET/POST/DELETE | `/client/machines` | Máquinas pessoais |
| POST | `/client/run` | Executar job |
| GET | `/client/jobs` | Histórico de jobs |
| GET | `/billing/status` | Saldo |
| POST | `/billing/purchase` | Comprar créditos |
| GET | `/billing/transactions` | Histórico |

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 11-admin-panel.md](11-admin-panel.md) &nbsp; | &nbsp; [Próximo: 13-provider-mode.md ➡️](13-provider-mode.md)
