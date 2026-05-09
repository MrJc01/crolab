# 🔭 Visão Geral da Arquitetura de Produção (V3)

Colocar o Crolab em produção significa gerenciar um ecossistema distribuído, com alto grau de isolamento de processos e baixa latência de comunicação.

Em um ambiente de produção real, o Crolab abandona a "Prova de Conceito" que roda num terminal local e passa a adotar uma arquitetura corporativa conhecida como **Cloud P2P Orchestrator**.

## 🧩 Como os Componentes se Conectam

Abaixo está o fluxo vital de como um código escrito pelo usuário no navegador é processado nos servidores físicos.

1. **Frontend (React SPA Embutido):**
   O usuário acessa o site (ex: `https://app.crolab.io`). Esse site não está rodando num servidor Nginx ou Node.js tradicional. Ele é sacado diretamente da memória do binário executável em Go graças à nossa arquitetura **Go Embed**. O navegador carrega o React e estabelece uma conexão WebSocket persistente com a API (Porta `8080`).

2. **O Roteador Central (Go Gateway):**
   O binário único do Go recebe a requisição WebSocket através da classe `Hub` (em `internal/gateway`). Ele converte o tráfego que vem da internet para um formato seguro e decide para qual Máquina Virtual (Kernel) ele deve ir.

3. **Orquestração de Nós (Kubernetes + KVM):**
   Para cada sessão aberta pelo usuário, um **Custom Resource (CRD)** do Kubernetes é disparado (`CrolabKernel`). Esse operador solicita ao servidor Linux host (Bare Metal) a criação imediata de uma microVM.

4. **Isolamento de Segurança (Firecracker):**
   No Servidor Físico, a biblioteca `internal/sandbox` do Go intercepta a ordem e:
   - Aciona a API de Socket Unix do **Firecracker**.
   - Cria uma VM contendo Python/Node.
   - Restringe essa VM em um **Cgroup V2** para garantir que o cliente não roube CPU/Memória do vizinho.
   - Tranca a rede usando **Network Namespaces (ip netns)**.

5. **A Execução Bidirecional (ZeroMQ):**
   Com a máquina ligada (em 50ms), a conexão WebSocket do Go deságua em uma comunicação **ZeroMQ (REQ/REP)** hiper-veloz pela porta `5555`. O código processa dados científicos e devolve instantaneamente pro Go, que atualiza a tela do usuário.

6. **Soberania de Dados (MinIO/S3):**
   Para evitar que os dados fiquem atrelados a esses contêineres efémeros, os Notebooks (os arquivos `.ipynb` com os dados do usuário) são continuamente sincronizados em Background para um **Storage de Objetos da AWS (S3)** ou um cluster **MinIO** através do pacote `internal/storage`.

---
**Próximo Passo:** Antes de compilar, você precisa entender os requisitos físicos da infraestrutura em [01-INFRA-REQUISITOS.md](01-INFRA-REQUISITOS.md).
