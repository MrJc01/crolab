*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 02-comandos.md](02-comandos.md) &nbsp; | &nbsp; [Próximo: 04-arquitetura.md ➡️](04-arquitetura.md)
<hr>

# Configuração de Servidores e Prioridade

## Conceito de Prioridade

Cada servidor tem um número de prioridade (1 = mais alto). Quando você executa `crolab run` sem `--target`, o sistema:

1. Ordena todos os servidores por prioridade (menor número primeiro)
2. Tenta conectar no primeiro
3. Se falhar, tenta o próximo (failover automático)
4. Se todos falharem, mostra erro

## Providers

O campo `provider` é informativo mas importante para organização:

| Provider | Quando usar |
|---|---|
| `local` | Sua máquina pessoal ou rede local |
| `vastai` | Instância alugada na Vast.ai |
| `runpod` | Instância na RunPod |
| `aws` | Amazon EC2 / SageMaker |
| `gcp` | Google Cloud Compute |
| `azure` | Microsoft Azure |
| `crom` | Máquina oficial da Crom Cloud |

## Exemplo: Pool de Prioridades

```bash
# GPU pessoal (tenta primeiro — grátis)
crolab config add minha-gpu 192.168.1.10:4422 tok --provider local --priority 1

# Vast.ai barata (fallback)
crolab config add vast-a100 45.67.89.10:4422 tok --provider vastai --priority 2

# RunPod (mais caro, última opção)
crolab config add runpod-01 123.45.67.89:4422 tok --provider runpod --priority 3
```

Ao rodar `crolab run .`, o sistema tenta `minha-gpu` → `vast-a100` → `runpod-01`.

## Configuração de Slots

Cada node tem um limite de jobs simultâneos:

```bash
# Node aceita até 4 jobs em paralelo
crolab serve --slots 4
```

Se os slots estiverem cheios, novos jobs entram em fila (máximo 5 minutos de espera).

## Tokens de Autenticação

```bash
# Gerar token automático
crolab serve --generate-auth --port :4422
# Output: Token gerado: cl_a8f7b2c3d4e5...

# Usar token específico
crolab serve --token meu-token-personalizado
```

Sem `--token`, o node aceita qualquer conexão (não recomendado em produção).

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [⬅️ Anterior: 02-comandos.md](02-comandos.md) &nbsp; | &nbsp; [Próximo: 04-arquitetura.md ➡️](04-arquitetura.md)
