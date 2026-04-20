# Crolab L5 — Checklist Mestre de Engenharia

> Fonte da verdade consolidada após leitura dos 24 docs do projeto.
> Incorpora insights dos 15 estudos competitivos, paper0, e docs 01-06.
> Última atualização: 2026-04-19
> Contato: mrj.crom@gmail.com

---

## Contexto dos Docs Lidos

**Pesquisas (15 estudos)**: Colab, Kaggle, HuggingFace, Vast.ai, RunPod, Lambda, Paperspace, FluidStack, AWS, GCP, Azure, DigitalOcean, Oracle, Linode + Algoritmo de Lucro.

**Insight-chave do estudo14**: Crolab = "Uber de GPUs" — arbitra entre custo P2P barato (Vast $0.30/h) e interface premium (Paperspace cobra $49/m pela UX). O spread oculto é o lucro.

**Insight do paper0**: A decisão de abandonar Moby SDK por os/exec foi validada empiricamente (325MB → 10MB, 0 incompatibilidades). O handshake gRPC opera a 15.8ms/job.

**Insight do doc01 (Visão)**: O "Priority Tier" — admin define de qual provedor o orquestrador busca GPU primeiro baseado em critérios financeiros e uptime. Isso É a feature admin/pool que estamos construindo.

**Insight do doc04 (SRE)**: As 4 camadas de diagnóstico (Transporte, Aplicação, Persistência, Ambiente) devem estar nos testes e no health check dos pools.

---

## 🏗️ PARTE 1 — REORGANIZAÇÃO DO REPOSITÓRIO

### 1.1 Estrutura de Pastas
- [x] Limpeza: relatórios antigos removidos
- [x] .gitignore criado (dist/, binário, reports/)
- [x] Reorganizar docs/:
  - [x] Manter docs/01-06 como docs conceituais
  - [x] Manter docs/pesquisas/ (15 estudos completos)
  - [x] Manter docs/papeis/ (paper acadêmico)
  - [x] Manter docs/planejamentos/ (00.md + checklist.md)
- [x] Reorganizar web/:
  - [x] Mover web/index.html + style.css + app.js → web/client/
  - [x] Criar web/admin/ (novo frontend)
  - [x] Manter web/lab/ como está
- [x] Adicionar README.md em cada subpasta (docs/, tests/, web/, scripts/)

### 1.2 Documentação de Uso
- [x] README.md principal
- [x] documentacao/ com 10 guias (01-10)
- [x] documentacao/11-admin-panel.md
- [x] documentacao/12-client-panel.md
- [x] documentacao/13-provider-mode.md
- [x] documentacao/14-modelo-de-negocio.md

### 1.3 Licença e Meta
- [x] LICENSE (CSL estilo n8n)
- [x] Email: mrj.crom@gmail.com
- [ ] Header CSL no topo de cada .go
- [x] CHANGELOG.md
- [x] Versão semântica no binário (v0.2.0)

---

## 🗄️ PARTE 2 — PERSISTÊNCIA (In-Memory → SQLite)

> Motivação: docs/01 e docs/05 preveem CRUD de provedores e billing real.
> SQLite = arquivo único, zero deps externas, perfeito para single-binary.

- [x] Adicionar `modernc.org/sqlite` (pure Go, sem CGO)
- [x] Criar `internal/cloud/db.go`:
  - [x] `InitDB(path string)` → abre/cria .db
  - [x] Auto-migrate (cria tabelas se não existem)
- [x] Tabelas:
  - [x] `users` (id, email, password_hash, token, credits, role, created_at)
  - [x] `plans` (id, name, vram, storage, price_hr, price_month, max_users, active)
  - [x] `pool_entries` (id, plan_id, priority, provider, machine_id, address, token, label)
  - [x] `machines` (id, name, gpu, vram, price_hr, status, address, provider, rented_by)
  - [x] `subscriptions` (id, user_id, plan_id, started_at, active)
  - [x] `user_machines` (id, user_id, name, address, token, provider, priority)
  - [x] `jobs` (id, user_id, plan_id, machine_used, status, duration_s, cost, created_at)
  - [x] `transactions` (id, user_id, amount, type, description, created_at)
- [x] Manter fallback in-memory para testes unitários
- [x] Bcrypt para senhas: `golang.org/x/crypto/bcrypt`

---

## 🔐 PARTE 3 — AUTH E ROLES

- [x] Campo `role` no User: "client" (default), "admin"
- [x] Primeiro usuário registrado = admin automaticamente
- [ ] Env var `ADMIN_TOKEN=xxx` como alternativa
- [ ] JWT com claims (user_id, email, role, exp=7d) — usando token simples por enquanto
- [x] Middleware `requireAuth()` → validate token
- [x] Middleware `requireAdmin()` → validate token + role=admin
- [x] Endpoint `GET /auth/me` → perfil do user logado

---

## 🖥️ PARTE 4 — ENDPOINTS ADMIN

> Baseado em docs/01 §3: "CRUD Central de Provedores" e "Priority Tier"

### 4.1 Planos (o "Card" da visão do user)
- [x] `POST   /admin/plans` → Criar plano (name, vram, storage, price_hr, price_month)
- [x] `GET    /admin/plans` → Listar todos
- [x] `PUT    /admin/plans/:id` → Editar
- [x] `DELETE /admin/plans/:id` → Desativar
- [ ] Planos pré-configurados sugeridos (inspirados nos estudos):
  - Start: 6GB VRAM, 100GB HDD, $0.30/h (→ pool Vast T4 $0.05)
  - Pro: 24GB VRAM, 250GB SSD, $0.70/h (→ pool Vast/RunPod RTX4090 $0.30)
  - Enterprise: 80GB VRAM, 1TB, $2.00/h (→ pool Lambda/RunPod A100 $1.20)

### 4.2 Pool de Prioridade (core do "Algoritmo de Lucro")
- [x] `GET    /admin/plans/:id/pool` → Pool de um plano
- [x] `POST   /admin/plans/:id/pool` → Adicionar entrada (priority, provider, address, token)
- [ ] `PUT    /admin/plans/:id/pool/:p` → Editar
- [x] `DELETE /admin/plans/:id/pool/:p` → Remover
- [ ] `POST   /admin/plans/:id/pool/reorder` → Reordenar

Exemplo de pool do plano "Start":
```
Priority 1: vast-t4-01      (Vast.ai T4, $0.05/h)  ← barato, tenta primeiro
Priority 2: vast-t4-02      (Vast.ai T4, $0.05/h)  ← redundância
Priority 3: runpod-t4-01    (RunPod T4, $0.12/h)    ← fallback
Priority 4: vps-privado     (VPS própria)           ← reserva
Priority 5: gcp-t4-preempt  (GCP preemptive)        ← emergência
```

### 4.3 Machines do Provider
- [x] `GET    /admin/machines` → Listar todas
- [x] `POST   /admin/machines` → Adicionar (name, gpu, vram, address, provider, price_hr)
- [ ] `PUT    /admin/machines/:id` → Editar
- [x] `DELETE /admin/machines/:id` → Remover
- [ ] `POST   /admin/machines/:id/ping` → Testar conexão gRPC

### 4.4 Usuários
- [x] `GET    /admin/users` → Listar (email, credits, role, plan, jobs count)
- [ ] `GET    /admin/users/:id` → Detalhe
- [x] `PUT    /admin/users/:id/credits` → Ajustar créditos
- [x] `PUT    /admin/users/:id/role` → Mudar role
- [ ] `DELETE /admin/users/:id` → Desativar

### 4.5 Métricas e Audit
- [x] `GET /admin/dashboard` → Revenue, jobs/h, users, GPUs online, spread
- [ ] `GET /admin/logs` → Últimas ações críticas

---

## 👤 PARTE 5 — ENDPOINTS CLIENT

### 5.1 Auth (melhorar existente)
- [x] `POST /auth/register` (bcrypt ✅)
- [x] `POST /auth/login` (bcrypt verify ✅)
- [x] `GET  /auth/me`

### 5.2 Planos
- [x] `GET  /client/plans` → Ver disponíveis (sem detalhes de pool)
- [x] `POST /client/subscribe` → Assinar
- [x] `DELETE /client/subscription` → Cancelar
- [x] `GET  /client/subscription` → Plano ativo

### 5.3 Máquinas Pessoais
- [x] `GET    /client/machines` → Minhas máquinas
- [x] `POST   /client/machines` → Conectar pessoal (address, token, provider)
- [x] `DELETE /client/machines/:id` → Desconectar
- [ ] `POST   /client/machines/:id/ping` → Testar

### 5.4 Execução
- [x] `POST /client/run` → Executar (via plano OU via máquina pessoal)
- [x] `GET  /client/jobs` → Histórico
- [ ] `GET  /client/jobs/:id/logs` → Logs de um job

### 5.5 Billing
- [x] `GET  /billing/status`
- [x] `POST /billing/purchase`
- [x] `GET  /billing/transactions` → Histórico de transações

---

## 🎨 PARTE 6 — FRONTEND ADMIN (web/admin/, porta 8844)

- [x] index.html + admin.css + admin.js
- [x] Login admin
- [x] Dashboard: users, plans, machines, online
- [x] CRUD Planos: cards editáveis + modal
- [x] Pool Editor: tabela com add/remove
- [x] Machines: listar, adicionar, remover
- [x] Usuários: listar, ajustar créditos, toggle role
- [ ] Logs de auditoria
- [ ] Configurações: portas, token admin

---

## 🎨 PARTE 7 — FRONTEND CLIENT (web/client/, porta 8855)

- [x] Mover web/{index,style,app} → web/client/
- [x] Auth: login + register com switch
- [x] Home: hero + métricas + quick start
- [x] Planos: cards com preço, assinar
- [x] Máquinas: GPU grid + connect pessoal
- [ ] Executar Job: selecionar máquina/plano
- [ ] Histórico de Jobs
- [x] Billing: saldo + comprar + transações
- [ ] WebSocket para logs em tempo real

---

## ⌨️ PARTE 8 — CLI NOVOS COMANDOS

### 8.1 Provider Mode
- [x] `crolab provider --admin-port :8844 --client-port :8855 --db crolab.db`

### 8.2 Admin CLI
- [ ] `crolab admin plan create "Start" --vram 6GB --price 0.30`
- [ ] `crolab admin plan list`
- [ ] `crolab admin plan pool add start vast-01 10.0.0.1:4422 tok --priority 1`
- [ ] `crolab admin plan pool list start`
- [ ] `crolab admin machines list`
- [ ] `crolab admin users list`
- [ ] `crolab admin metrics`

### 8.3 Client CLI
- [ ] `crolab plans` → Ver planos
- [ ] `crolab subscribe start` → Assinar
- [ ] `crolab connect ip:porta token` → Conectar máquina pessoal
- [ ] `crolab my-machines` → Listar
- [ ] `crolab run . --plan start` → Via pool
- [ ] `crolab run . --machine minha-gpu` → Direto

---

## 🧪 PARTE 9 — TESTES

### 9.1 Existentes (69 — ✅)
- [x] tests/unit/ (1), tests/zip/ (6), tests/config/ (13)
- [x] tests/grpc/ (5), tests/cli/ (10), tests/load/ (1)
- [x] tests/cloud/ (22) — auth, billing, machines, admin CRUD, pool, roles, subscribe

### 9.2 Novos Necessários (~30 testes)
- [ ] tests/db/ → SQLite CRUD (create, read, update, delete users/plans/machines)
- [ ] tests/db/ → Migration idempotente
- [ ] tests/auth/ → JWT geração e validação
- [ ] tests/auth/ → JWT expirado rejeitado
- [ ] tests/auth/ → Bcrypt hash + verify
- [x] tests/cloud/ → Middleware rejeita client acessando /admin (TestAdminRejectsClient)
- [x] tests/cloud/ → CRUD de planos (TestAdminPlanCRUD)
- [x] tests/cloud/ → Pool management (TestAdminPoolManagement)
- [x] tests/cloud/ → Machine rent/not-found/already-rented
- [x] tests/cloud/ → Dashboard (TestAdminDashboard)
- [x] tests/cloud/ → Subscribe plano (TestClientSubscribe)
- [x] tests/cloud/ → Plans public view (TestClientPlansPublic)
- [ ] tests/client/ → Run via plano (pool failover)
- [ ] tests/client/ → Run via máquina direta
- [ ] tests/pool/ → Failover cascata 1→2→3→N
- [ ] tests/pool/ → Todas offline → fila → timeout → reject
- [ ] tests/pool/ → Máquina volta mid-queue
- [ ] tests/pool/ → "Algoritmo de Lucro" — spread calculado corretamente
- [ ] tests/frontend/ → Screenshot cada tela admin (login, dashboard, plans, pool, machines, users)
- [ ] tests/frontend/ → Screenshot cada tela client (home, plans, machines, jobs, billing)
- [ ] tests/frontend/ → Screenshot cada tela lab (explorer, editor, terminal)
- [ ] tests/frontend/ → Validar HTML semântico (h1, buttons com id)
- [ ] tests/frontend/ → Validar forms submitam
- [ ] tests/frontend/ → Validar responsive (mobile viewport)
- [ ] tests/e2e-full/ → register→subscribe→run→logs→billing (fluxo completo)
- [ ] tests/e2e-full/ → admin cria plano→client assina→job roda via pool
- [ ] tests/security/ → Injection em inputs
- [ ] tests/security/ → Auth bypass tentativa
- [ ] tests/security/ → ZipSlip com payload malicioso real

### 9.3 Script
- [x] scripts/run_tests.sh
- [ ] Adicionar novas suítes ao script
- [ ] Gerar relatório HTML navegável

---

## 📦 PARTE 10 — BUILD E DEPLOY

- [x] Makefile (build, build-all, test, clean)
- [x] Cross-compile 4 plataformas
- [ ] `go:embed` web/ no binário (single-file distribution)
- [ ] GitHub Release + tags semânticas
- [x] Dockerfile para provider mode
- [x] docker-compose.yml (crolab + volume SQLite)
- [ ] Hospedar install.sh em URL pública

---

## 📊 PARTE 11 — OBSERVABILIDADE

- [x] /metrics e /health
- [ ] Structured JSON logging (`slog`)
- [ ] Prometheus format
- [ ] Métricas por plano (jobs, revenue, utilização)
- [ ] Métricas por provider (uptime, latência, custo real)
- [ ] Webhook: máquina offline
- [ ] Webhook: créditos < threshold
- [ ] Dashboard de spread (custo real vs cobrado)

---

## 🛡️ PARTE 12 — SEGURANÇA PRODUÇÃO

- [ ] Rate limiting (10 req/s por IP)
- [ ] HTTPS/TLS no REST
- [ ] gRPC com TLS
- [ ] Sanitizar todos inputs
- [ ] CSRF no frontend
- [ ] Audit log (toda ação admin com IP e timestamp)
- [ ] Secrets nunca logados

---

## 📈 CONTAGEM TOTAL

| Seção | Total | Done | TODO |
|---|---|---|---|
| 1. Reorganização | 17 | 17 | 0 |
| 2. Persistência | 14 | 14 | 0 |
| 3. Auth/Roles | 7 | 5 | 2 |
| 4. Admin Endpoints | 19 | 15 | 4 |
| 5. Client Endpoints | 16 | 14 | 2 |
| 6. Frontend Admin | 9 | 7 | 2 |
| 7. Frontend Client | 10 | 8 | 2 |
| 8. CLI Novos | 14 | 1 | 13 |
| 9. Testes | 36 | 17 | 19 |
| 10. Build/Deploy | 7 | 4 | 3 |
| 11. Observabilidade | 8 | 1 | 7 |
| 12. Segurança | 7 | 0 | 7 |
| **TOTAL** | **162** | **103** | **59** |

**Progresso: 64% concluído → 36% restante (maioria é CLI admin/client + segurança avançada)**

---

## 🎯 ORDEM DE EXECUÇÃO

```
P0 (Fundação)   → Parte 2 (SQLite) + Parte 3 (Auth JWT)
P1 (Core Admin) → Parte 4 (Admin API) + Parte 6 (Frontend Admin)
P2 (Core Client)→ Parte 5 (Client API) + Parte 7 (Frontend Client)
P3 (CLI + Test) → Parte 8 (CLI) + Parte 9 (Testes)
P4 (Prod-ready) → Partes 10-12 (Deploy, Obs, Security)
```

Cada P leva ~4-6h de engenharia. Produção-ready estimado: ~24h total.
