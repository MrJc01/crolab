# /tests

Suíte de testes automatizados do Crolab.

## Estrutura

| Pasta | Testes | O que testa |
|---|---|---|
| `unit/` | 1 | Helpers básicos |
| `zip/` | 6 | Compressão/descompressão ZIP |
| `config/` | 13 | Configuração de servidores |
| `grpc/` | 5 | Protocolo gRPC de jobs |
| `cli/` | 10 | Comandos CLI |
| `cloud/` | 22 | Auth, billing, admin CRUD, roles, pool |
| `e2e/` | 1 | Fluxo completo admin→client (15 steps) |
| `load/` | 1 | Teste de carga |

## Rodar

```bash
# Todos
go test ./tests/...

# Específico
go test -v ./tests/cloud/...

# Com relatório
bash scripts/run_tests.sh
```
