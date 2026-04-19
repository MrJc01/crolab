# Estudo Competitivo: Lambda Labs

## 1. Arquitetura (Hardware como Serviço de Base)
A Lambda Labs difere do modelo Vast.ai por possuir foco hardcore em **Hardware Físico Próprio** e montagem de *Datacenters Dedicados para AI*. A arquitetura central foge um pouco do "Container isolado efêmero" e empurra o modelo `1-Click Clusters`. Trata-se de infraestrutura de Altíssimo Desempenho usando rede NVLink e InfiniBand desenhada para orquestrações Enterprise (ex: Kubernetes ou *Slurm*, o gestor clássico de Supercomputadores focado em HPC).

## 2. Média de Valor/Custo
A faturabilidade é feita por On-Demand ou Contratos Reservados (muito mais baratos que as concorrentes Azure e GCP). Uma placa robusta A10 começa por `$0.75/hora`. A lendária H100 custa aproximadamente `$3.40 a $4.00`. Fator de ouro: Zero taxa de transferência interna e **Sem Custo de Egress**, barateando brutalmente pipelines AI.

## 3. Como Funciona na Prática
O Desenvolvedor aluga um nó e a Lambda te atira o *Lambda Stack*: Ambiente pré-mastigado onde Ubuntu, Cuda, Drivers e PyTorch já caem redondos sem as infinitas dores de cabeça do WSL2 do windows.

## 4. Visão Crolab (Cluster Slurm P2P)
As corporações de Altíssimo Escopo (Foundation Models) amam Kubernetes e o velho Slurm. 
O Agente Binário CLI do Crolab deve prever uma injeção Slurm! Nossa Trindade Node pode se ramificar pra submeter Jobs orquestrados num cluster nativo local dentro dos H100 Locados deles, abstraindo Kubernetes pro Usuário sem perder conectividade com nossa interface Cloud Web App futura. Lucro na estabilidade do hardware alheio.
