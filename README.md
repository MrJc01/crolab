# ⚡ Crolab V3 — The Cloud P2P Data Orchestrator

Bem-vindo ao repositório oficial do **Crolab**, a plataforma definitiva que democratiza o processamento em nuvem. Se você já usou o **Google Colab** ou o **Jupyter Notebook**, o Crolab é exatamente isso, mas desenhado para ser infinitamente escalável, seguro e conectado em uma malha P2P (ponto a ponto).

O objetivo do Crolab é revolucionar o custo e a forma de rodar **Modelos de Linguagem (IA)** e **Data Science**, conectando de forma segura quem precisa de processamento (Clientes) com quem tem Hardware sobrando (Provedores).

---

## 🧠 Como o Crolab Funciona? (Explicado de Forma Simples)

Imagine um restaurante gigantesco:
- **O Cliente (React Frontend):** É o navegador do usuário. Ele possui um cardápio bonito onde ele digita o código (o pedido). Para garantir que o pedido nunca se perca se acabar a luz, ele tem um caderninho próprio que anota tudo em tempo real (Persistência Local-First com IndexedDB).
- **O Garçom Maestro (Go Gateway):** É o cérebro central rápido e que nunca dorme. Escrito em **Go**, ele usa "túneis de rádio" instantâneos (WebSockets) para pegar o pedido do cliente e levar até a cozinha em milissegundos.
- **As Cozinhas Isoladas (Firecracker MicroVMs):** Onde a mágica acontece. Mas temos um problema: e se um cliente pedir uma "bomba" em vez de um prato? Para proteger o restaurante, cada pedido é preparado dentro de uma **Caixa Forte blindada (Sandbox)**. O Crolab cria uma Máquina Virtual inteira do zero em apenas 50 milissegundos para rodar o código do cliente. Terminou? A máquina é destruída.
- **A Janela da Cozinha (ZeroMQ):** Para o cozinheiro passar o prato de volta para o garçom com rapidez absoluta, usamos uma tecnologia de alta performance de dados chamada **ZeroMQ**.

### 🌟 Principais Tecnologias (Stack)
- **Frontend Interativo:** React, Vite, Zustand e Monaco Editor (O mesmo editor do VS Code).
- **Backend & Gateway:** Go (Golang) lidando com concorrência massiva via WebSockets.
- **Segurança (Sandbox):** Instâncias Firecracker, isoladas cirurgicamente por Cgroups V2 (que limita CPU/RAM) e Network Jails (que impede acesso à rede interna).
- **Banco de Dados (Nuvem):** MinIO / Amazon S3 para guardar os cadernos Jupyter (.ipynb) de forma vitalícia.
- **Orquestração e Deploy:** Kubernetes e GitHub Actions automatizando tudo.

---

## 🚀 Como Fazer o Download e Instalar (Gold Master Release)

O Crolab convergiu todo seu motor para um **Binário Autônomo Único**. Isso significa que a interface gráfica (React) está **embutida** dentro do executável Go. Você não precisa instalar Node.js ou configurar servidores web complicados.

Você pode instalá-lo de forma autônoma pela linha de comando em Servidores Nuvem Linux ou Local Mac via:
```bash
curl -sSL https://crolab.crom.run/install | bash
```

Se desejar instalação via binários diretos `.zip` / `.exe`:
1. **Vá na aba "Releases"** no canto direito deste repositório no GitHub.
2. Faça o download da versão Zip do seu Sistema — ex: `crolab-linux-amd64.zip`.
3. Extraia o executável e digite no terminal para inicializar a Trindade Orquestradora:
   `crolab provider start`

---

## 📚 Navegando Pela Documentação

Criamos uma Árvore Hierárquica completa para você navegar linearmente e explorar este software a fundo de modo seguro:

### 🌐 Começando
*   [01. Instalação e Arquitetura Prévia](documentacao/01-instalacao.md)
*   [02. CLI e Comandos Padrão](documentacao/02-comandos.md)
*   [03. Configuração de Variáveis Sensíveis](documentacao/03-configuracao.md)

### ⚙️ Engenharia Backend e Segurança SRE
*   [04. Arquitetura da Plataforma e Gateway](documentacao/04-arquitetura.md)
*   [05. Cloud API e Webhooks n8n](documentacao/05-cloud-api.md)
*   [08. Segurança Paranoica (Firecracker e Cgroups)](documentacao/08-seguranca.md)

### 🎨 Interface e Economia
*   [06. Lab Jupyter Environment (React)](documentacao/06-lab.md)
*   [13. Modo Provedor (Alugar a GPU)](documentacao/13-provider-mode.md)
*   [14. Modelo de Transações / Billing](documentacao/14-modelo-de-negocio.md)

---

> 💡 *A navegação na pasta `/documentacao` está automatizada. Ao clicar em qualquer assunto, você encontrará botões padronizados no topo e no final da página para [🔙 Voltar ao Índice] ou [Próximo Capítulo ➡️]. Aproveite seus estudos!*
