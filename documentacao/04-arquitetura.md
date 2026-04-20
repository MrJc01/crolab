*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 03-configuracao.md](03-configuracao.md) &nbsp; | &nbsp; [Próximo: 05-cloud-api.md ➡️](05-cloud-api.md)
<hr>

# Arquitetura do Sistema

## O Binário Trindade

O Crolab é um único binário Go que opera em 3 modos:

```
┌──────────────────────────────────────────────┐
│              CROLAB BINARY                    │
├──────────────┬──────────────┬────────────────┤
│   CLI Mode   │  Node Mode   │  Cloud Mode    │
│  (run, cfg)  │  (serve)     │ (cloud-serve)  │
├──────────────┼──────────────┼────────────────┤
│ Empacota zip │ Recebe gRPC  │ REST API       │
│ Envia gRPC   │ Docker exec  │ Auth/Billing   │
│ Tail logs    │ Stream logs  │ Frontend web   │
└──────────────┴──────────────┴────────────────┘
```

## Fluxo de Execução de um Job

```
1. CLI empacota diretório → ZIP (bytes)
2. CLI conecta ao Node via gRPC
3. CLI envia JobRequest (image, command, payload)
4. Node extrai ZIP em /tmp/workspace-{uuid}
5. Node executa: docker run --rm -v workspace:/workspace image cmd
6. Docker stdout/stderr → pipe → chan string
7. Chan string → gRPC StreamLogs → CLI
8. CLI imprime logs em tempo real
9. Node limpa workspace temporário
```

## Componentes

### `internal/cli/` — Cliente

- **config.go**: CRUD de servidores com Viper (YAML), ordenação por prioridade, failover
- **job_push.go**: ZipDir (comprime), SubmitJob (gRPC), tailLogs (stream)

### `internal/node/` — Daemon

- **server.go**: gRPC server, semáforo de slots, fila com timeout, auth interceptors
- **docker_runner.go**: `os/exec` → Docker, pipes stdout/stderr para channel
- **gpu.go**: Detecção via `nvidia-smi`, flags `--gpus` para Docker
- **metrics.go**: HTTP `/metrics` e `/health` na porta 9090

### `internal/cloud/` — Backend SaaS

- **server.go**: REST API (net/http), CORS, serve frontend estático

### `internal/lab/` — Editor Web

- **server.go**: API de arquivos (list/read/save), WebSocket para execução

### `internal/tui/` — Interface Terminal

- **monitor.go**: BubbleTea dashboard com ping, tabs, formulário
- **selector.go**: Seletor interativo de servidor

## Protocolo gRPC

```protobuf
service CrolabService {
  rpc SubmitJob(JobRequest) returns (JobResponse);
  rpc StreamLogs(LogRequest) returns (stream LogMessage);
}
```

O payload (código zipado) vai dentro do `JobRequest.payload` (bytes).

## Segurança

```
Cliente                          Servidor
  │                                │
  │── metadata: authorization ───→ │
  │                                │── TokenAuthInterceptor
  │                                │── Valida token
  │                                │── Se OK: processa
  │                                │── Se não: codes.Unauthenticated
```

## Multi-Tenancy

```
Job chega → Semáforo livre? ──→ Executa (Docker --cpus --memory)
                 │
                 └─ Cheio? → Fila (max 5min) → Timeout? → ResourceExhausted
```

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 03-configuracao.md](03-configuracao.md) &nbsp; | &nbsp; [Próximo: 05-cloud-api.md ➡️](05-cloud-api.md)
