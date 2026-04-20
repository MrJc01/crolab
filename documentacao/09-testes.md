*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 08-seguranca.md](08-seguranca.md) &nbsp; | &nbsp; [Próximo: 10-deploy.md ➡️](10-deploy.md)
<hr>

# Testes

## Rodar Todos os Testes

```bash
bash scripts/run_tests.sh
```

Saída:
```
═══════════════════════════════════════
  CROLAB — SUÍTE COMPLETA DE TESTES
═══════════════════════════════════════
  FASE 0: BUILD
  ▸ Compilando binário... ✓ OK (22M)

  FASE 1: TESTES UNITÁRIOS
  ▸ ZipDir (compressão)...           ✓ PASS (1 teste)
  ▸ Zip/Unzip (roundtrip + edge)...  ✓ PASS (6 testes)
  ▸ Config CRUD (add/rm/sort/hash)...✓ PASS (13 testes)

  FASE 2: TESTES DE INTEGRAÇÃO
  ▸ gRPC (submit/auth/timeout)...    ✓ PASS (5 testes)
  ▸ CLI Completa (help/status/crud)..✓ PASS (10 testes)

  FASE 3: TESTES DE CARGA
  ▸ Stress 50x gRPC (chaos loop)...  ✓ PASS (1 teste)

  FASE 4: CLOUD API
  ▸ Cloud API (register/login/rent)..✓ PASS (11 testes)

  RELATÓRIO FINAL
  Taxa de sucesso: 100% (47 testes)
```

## Estrutura

```
tests/
├── unit/     → ZipDir básico
├── zip/      → Zip múltiplos arquivos, vazio, inválido, unzip, roundtrip
├── config/   → AddServer, update, remove, set-default, sort, hash, persistência
├── grpc/     → Submit, image/command, bad address, auth metadata, timeout
├── cli/      → Help, status, config CRUD, flags de serve/run/lab
├── load/     → 50x gRPC stress test
├── cloud/    → Register, login, billing, machines, rent
└── reports/  → Relatórios gerados (um por execução)
```

## Rodar Suíte Específica

```bash
go test -v -count=1 ./tests/config/...
go test -v -count=1 ./tests/cloud/... -timeout 30s
go test -v -count=1 ./tests/grpc/...
```

## Relatórios

Cada execução de `run_tests.sh` gera um relatório em:
```
tests/reports/report_YYYYMMDD_HHMMSS.txt
```

O relatório contém output completo de cada teste, contagem de passa/falha, e resumo final.

## Adicionar Novos Testes

1. Crie uma pasta em `tests/` (se nova área)
2. Nomeie como `*_test.go`
3. Adicione o caminho no `scripts/run_tests.sh`

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 08-seguranca.md](08-seguranca.md) &nbsp; | &nbsp; [Próximo: 10-deploy.md ➡️](10-deploy.md)
