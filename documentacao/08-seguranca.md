*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 07-monitor.md](07-monitor.md) &nbsp; | &nbsp; [Próximo: 09-testes.md ➡️](09-testes.md)
<hr>

# Segurança

## Autenticação gRPC

Toda comunicação entre CLI e Node usa gRPC com autenticação via token:

```
CLI                              Node
 │                                │
 │── JobRequest + Authorization ──→ TokenAuthInterceptor
 │                                │── token == válido? → processa
 │                                │── token != válido? → Unauthenticated 401
```

### Ativando no Node

```bash
# Token manual
crolab serve --token meu-token-secreto

# Token gerado automaticamente
crolab serve --generate-auth
# Output: cl_a8f7b2c3d4e5f6g7h8i9...
```

### Usando no CLI

```bash
crolab config add meu-gpu 10.0.0.5:4422 meu-token-secreto
```

O token é enviado no header `authorization` de cada chamada gRPC.

## ZipSlip Guard

Ao descompactar payloads recebidos, o node verifica se caminhos extraídos tentam escapar do diretório workspace:

```go
if !strings.HasPrefix(destPath, filepath.Clean(dest)) {
    return fmt.Errorf("ZipSlip detectado: %s", name)
}
```

Isso previne ataques onde um ZIP malicioso tenta sobrescrever `/etc/passwd` ou similares.

## Isolamento Docker

Cada job roda em container isolado com:
- `--rm`: auto-limpeza após execução
- `--cpus 1.0`: limite de CPU
- `--memory 2g`: limite de memória
- Workspace temporário em `/tmp/crolab-job-{uuid}`

## Cloud API Auth

A API REST usa tokens Bearer no header `Authorization`:

```bash
curl -H "Authorization: abc123token" http://localhost:8844/billing/status
```

## Recomendações de Produção

1. **Sempre use `--tls-cert` e `--tls-key`** para habilitar o tunelamento criptografado no Control Plane e Data Plane. O tráfego sem TLS deve ser banido se operar fora de uma intranet local.
2. **Sempre use `--token`** no serve
3. **Nunca exponha a porta 4422** sem firewall ou TLS habilitado
4. **Use VPN** para conexões entre datacenters
5. **Rotacione tokens** periodicamente
6. **Monitore `/metrics`** para detectar anomalias (Prometheus)

## Segurança de Criptografia TLS O(1)

Ambas API REST Cloud e Nodes gRPC P2P aceitam as flags `--tls-cert` e `--tls-key` ativando Criptografia Assimétrica imediata de ponta-a-ponta, prevenindo interceptação (Sniffing) da payload de dados (Código, Tensores e Outputs). No cliente final, habilite o handshake gRPC validado com `--tls-rpc`.

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 07-monitor.md](07-monitor.md) &nbsp; | &nbsp; [Próximo: 09-testes.md ➡️](09-testes.md)
