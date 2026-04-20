*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [Próximo: 02-comandos.md ➡️](02-comandos.md)
<hr>

# Instalação e Configuração

## Requisitos

- Go 1.21+ (para compilar)
- Docker (para executar jobs)
- Linux ou macOS

## Compilar do Código Fonte

```bash
git clone https://github.com/crolab/core.git
cd crolab

# Build local
make build

# Cross-compile para todas as plataformas
make build-all
# Binários ficam em dist/
```

## Instalar em Máquina Remota

Execute na máquina remota:

```bash
curl -sSL https://crolab.crom.run/install | bash
```

O script:
1. Detecta OS e arquitetura
2. Baixa o binário correto
3. Instala em `/usr/local/bin/crolab`
4. Gera um token de autenticação
5. Mostra IP + hash para conectar

## Primeira Configuração

```bash
# Ver estado atual
./crolab status

# Adicionar seu primeiro servidor
./crolab config add meu-gpu 192.168.1.10:4422 token-aqui --provider local --priority 1

# Ver servidores configurados
./crolab config ls
```

## Arquivo de Configuração

O Crolab salva tudo em `~/.crolab/config.yaml`:

```yaml
default_server: meu-gpu
cloud_token: ""
servers:
  - name: meu-gpu
    address: "192.168.1.10:4422"
    token: "abc123"
    provider: local
    priority: 1
```

## Variáveis Importantes

| Campo | Descrição |
|---|---|
| `name` | Nome amigável do servidor |
| `address` | IP:porta do daemon gRPC |
| `token` | Token de autenticação |
| `provider` | Tipo: local, vastai, runpod, aws, gcp |
| `priority` | 1 = mais prioritário (usado no failover) |

<hr>
*[🔙 Voltar ao Hub Principal (README)](../README.md)* &nbsp; | &nbsp; [Próximo: 02-comandos.md ➡️](02-comandos.md)
