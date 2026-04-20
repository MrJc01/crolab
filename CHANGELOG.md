# Changelog

Todas as mudanças notáveis do projeto Crolab.

## [0.2.0] — 2026-04-19

### Adicionado
- **SQLite Persistence** — 8 tabelas (users, plans, pool_entries, machines, subscriptions, user_machines, jobs, transactions)
- **Bcrypt** — senhas hashadas com `golang.org/x/crypto/bcrypt`
- **Roles** — primeiro user = admin, demais = client
- **Middleware Auth** — `requireAuth()` e `requireAdmin()` para proteção de rotas
- **Admin API** — CRUD de planos, pool de prioridade, máquinas, usuários, dashboard, logs
- **Client API** — subscribe/unsubscribe, máquinas pessoais, jobs, billing com transações
- **Frontend Admin** — painel completo (login, dashboard, planos, pool editor, máquinas, usuários)
- **Frontend Client** — painel completo (auth, home, planos, GPUs, billing)
- **Provider Mode** — `crolab provider --admin-port :8844 --client-port :8855`
- **E2E Test** — fluxo completo de 15 passos (admin→client→subscribe→run→billing)
- **Documentação** — guias 11-14 (admin, client, provider, modelo de negócio)

### Modificado
- **server.go** — reescrito de in-memory para SQLite
- **cloud_test.go** — 22 testes novos (auth, billing, admin, client, roles)

## [0.1.0] — 2026-04-18

### Adicionado
- **Crolab Lab** — editor web com execução de scripts via WebSocket
- **47 testes** — unit, zip, config, grpc, cli, load, cloud
- **Documentação** — 10 guias (01-10) + README
- **Licença CSL** — estilo n8n, uso livre pessoal/educacional
- **Makefile** — build, build-all (4 plataformas), test, clean
- **Scripts** — `run_tests.sh` com relatórios
