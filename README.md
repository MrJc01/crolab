# ⚡ Crolab — Cloud GPU P2P & Data Orchestration

Bem-vindo ao repositório oficial do **Crolab**, um poderoso ambiente modular P2P projetado para revolucionar O custo de execução de Modelos de Linguagem e processamento científico denso (Machine Learning) conectando locadores e provedores de Hardware.

Ele embute nativamente ferramentas como:
- **Cloud Central & Client Dashboard** com interface Vanilla Glassmorphism P2P.
- **Kernel Sandbox Nativo** provendo container isolation SRE para Node/Python.
- **Micro-Gateway P2P (Node Provider Mode)** gerido em Go para emissão e recebimento de Payload em massa.

---

## 🚀 Como Fazer o Download e Instalar (Gold Master Release)

O Crolab convergiu todo seu motor (UI Web, Painel Administrativo CLI, Banco SQLite SRE e gRPC TLS Protocol) dentro de um binário autônomo único sem necessidade de frameworks pesados de retaguarda (*Go Embed Core*).

Você pode instalá-lo de forma autônoma pela linha de comando em Servidores Nuvem Linux ou Local Mac via:
```bash
curl -sSL https://crolab.crom.run/install | bash
```

Se desejar instalação via binários diretos `.zip` / `.exe`:
1. **Vá na aba "Releases"** no canto direito deste repositório no GitHub.
2. Faça o download da versão Zip do seu Sistema — ex: `crolab-linux-amd64.zip`.
3. Extraia o executável e digite no terminal para inicializar toda a Trindade Orquestradora na Porta Web/API em segundos:
   `crolab provider start`

Criamos uma Árvore Hierárquica completa para você navegar linearmente e explorar este software a fundo de modo seguro, podendo ir e voltar entre os tópicos a qualquer instante.

### 🌐 Começando
*   [01. Instalação e Arquitetura Prévia](documentacao/01-instalacao.md)
*   [02. CLI e Comandos Padrão](documentacao/02-comandos.md)
*   [03. Configuração de Variáveis Sensíveis](documentacao/03-configuracao.md)

### ⚙️ Engenharia Backend e Provedores
*   [04. Arquitetura da Plataforma](documentacao/04-arquitetura.md)
*   [05. Cloud API Gateway](documentacao/05-cloud-api.md)
*   [07. Logs e Monitoramento em Tempo Real](documentacao/07-monitor.md)
*   [08. Segurança Global e Chaves Criptadas](documentacao/08-seguranca.md)

### 🎨 Interfaces Gráficas
*   [06. Lab Jupyter Environment](documentacao/06-lab.md)
*   [11. Hub de Administração](documentacao/11-admin-panel.md)
*   [12. Portal do Cliente Colab-Style](documentacao/12-client-panel.md)

### 💰 Economia P2P e Ecossistema
*   [13. Modo Provedor (Alugar a GPU)](documentacao/13-provider-mode.md)
*   [14. Modelo de Transações / Billing](documentacao/14-modelo-de-negocio.md)

### 👩‍💻 Integração Avançada e Qualidade
*   [09. Ecossistema Forense e Testes E2E](documentacao/09-testes.md)
*   [10. Deploying Manual](documentacao/10-deploy.md)
*   [15. Usando SDK Remoto para Apps Externos](documentacao/15-sdk.md)

---

> 💡 *A navegação na pasta `/documentacao` está automatizada. Ao clicar em qualquer assunto, você encontrará botões padronizados no topo e no final da página para [🔙 Voltar ao Índice] ou [Próximo Capítulo ➡️]. Aproveite seus estudos!*

