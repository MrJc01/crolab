# 📦 Build e Empacotamento do Core (Go Embed)

A grande vantagem arquitetural do Crolab V3 é não exigir a instalação complexa de `NodeJS` + `Nginx` + `Go` no servidor de produção final.

O Crolab adota a filosofia do **Binário Único Autônomo** através da diretiva `//go:embed` nativa do compilador Go. Ele literalmente engole o site (SPA React) e insere tudo num executável de poucos megabytes.

## Como Realizar o Build Manualmente

Se você não for usar o CI/CD do GitHub Actions, siga os passos abaixo em seu ambiente de desenvolvimento.

### 1. Compilar o Frontend (React / Vite)
Entre na pasta do frontend, instale as dependências e faça a transpilação para HTML/JS estático.

```bash
cd web/frontend
npm install
npm run build
```

*(Isso vai gerar uma nova pasta chamada `dist/` contendo todos os assets compactados).*

### 2. Transportar para o Contexto do Go
O pacote `embed` do Go não pode ler diretórios acima da própria raiz do arquivo Go que declara a variável por motivos de segurança do compilador.
Portanto, a pasta recém-criada precisa ser movida para dentro do módulo `internal/web/`:

```bash
# Estando na raiz do repositório
mkdir -p internal/web/dist
cp -r web/frontend/dist/* internal/web/dist/
```

### 3. Compilação Cruzada (Cross-Compilation)
Agora, vamos compilar o Go. Se você estiver num Mac ou Windows, precisa dizer ao Go que o destino final é o Linux Host (Bare Metal) arquitetura AMD64 (ou ARM64 dependendo do seu hardware).

```bash
# Estando na raiz do repositório
export GOOS=linux 
export GOARCH=amd64 

# Tira toda a tabela de depuração (-s -w) para deixar o arquivo menor e compila:
go build -ldflags="-s -w" -o crolab-linux-amd64 cmd/crolab/main.go
```

Pronto. O arquivo `crolab-linux-amd64` agora é tudo o que você precisa. Ele carrega consigo o roteador WebSocket, os middlewares de métricas do Prometheus, as bibliotecas do ZeroMQ e todo o site do React.

## Entregando para o Servidor Host
Você pode simplesmente transferir esse único arquivo via SCP para o seu servidor bare metal de produção:

```bash
scp crolab-linux-amd64 root@ip-do-servidor:/usr/local/bin/crolab
```

No servidor host, você apenas digitaria `crolab cloud-serve start` para rodar todo o complexo orquestrador em modo daemon!

---
**Próximo Passo:** O binário roda, mas para suportar a criação das máquinas isoladas, precisamos preparar as raízes de rede e do HD do Firecracker. Leia [03-FIRECRACKER-SETUP.md](03-FIRECRACKER-SETUP.md).
