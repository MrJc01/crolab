# Crolab: Checklist Sistêmico de Tarefas e Backlog do MVP

Ao iterar contra um modelo pragmático para o Crolab, o cronograma a seguir prioriza faturamento ágil da empresa via Orquestração e Conectividade Direta nativa do usuário usando puro Docker.

## Fase 1: MVP do Binário Node-CLI (Isolado e Pessoal)
A prova de conceito inicial não envolve sequer um Backend Web Central. Foca 100% no motor operando de maneira avulsa.

- [ ] **Daemon Base (Node):** Módulo `cmd/crolab-node` que implemente um servidor RPC (Ouvinte nativo). Tem a função primordial de interceptar metadados, startar contêiner de imagem Docker Padrão e manter o lock do processo.
- [ ] **Conexão Livre (Sem Middleware Crom):** Possibilitar que a configuração inicial do cliente conecte localmente via IP próprio. Comando `crolab config set-server IP:PORTA`.
- [ ] **Empacotamento Subjetivo:** Construção da função "Job Submission": Comprime a pasta local em .zip, envia os binários via *Data Stream*, o Node extrai, repassa o Volume para o Docker e executa a linha designada.
- [ ] **Túnel de Telemetria Contínua:** Mapeamento em tempo real (Tail/Follow) do Log Console via RPC do `crolab-node` pro Node terminal de entrada, para sobreviver ao Fechamento abrupto das telas. E no término salva tudo e retorna.

## Fase 2: Plataforma de Orquestração Crom (O Negócio "Cloud")
Completado o MVP onde o PC conecta com qualquer host passivo, iniciam-se os controladores SaaS.

- [ ] **O Roteador Multi-Provedores:** Backend em Go acoplado ao Identity Hub para expor `crolab.dev/api/v1/jobs`.
- [ ] **Painel de CRUD da Nuvem:** Desenvolvimento Web de Backoffice mapeando Vast.ai, GCP e AWS. Configuração de "Priority Tier" (De qual provedor o orquestrador buscará a GPU primeiro a depender de critérios financeiros e Uptime).
- [ ] **Billing Engine (Créditos Crom):** Camada de faturamento que calcula o multiplicador/markup de lucro (Ex: Instância Vast custa $0.05 a hora, repassada pro usuário como $0.06 abatido via créditos/stripe).
- [ ] **Brokerage Comunitário:** O `crolab-node` ganha opção `crolab-node register-provider --wallet=UUID`. Cadastrando o node no Back-Office da CLI publicando seu host online para "ganhar a fatia" dos treinos P2P sob as asas do sistema.

## Fase 3 (Backlog Tardio do Estaleiro)

- [ ] **Integração do Crompressor:** Refatoração do sistema de Transferência "Zip" para envio binário assinado Delta (Vide docs `03_V2`).
- [ ] **Interface Visual Web Baseada no CLI:** WebApp mapeando métricas ao invés do terminal cru.
- [ ] **Plugin Bridge:** Estudar proxy para rodar instâncias "grátis" logadas em plataformas alheias do usuário (Ex: Contas Colab antigas).
