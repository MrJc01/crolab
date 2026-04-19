# Estudo Competitivo: Amazon Web Services (EC2)

## 1. Arquitetura Base (Hipervisores Monolíticos Legacy)
A AWS foi a detentora da internet mundial por décadas operando Infraestrutura como Serviço (IaaS). O coração deles rege nas arquiteturas EC2 impulsionadas organicamente por Hipervisores de Segurança proprietários como o AWS Nitro.

## 2. Média de Valor/Custo
Praticamente as faturas **mais onerosas do setor computacional L3 isolado**. Custos atrelados são surreais pra faturamento. Máquinas da Geração `p4d` chegam em patamares abusivos (U$20 a U$40 hora), as opções de Data Egress Tax cobram transferências oceânicas e discos SSD alocáveis via Amazon EBS.

## 3. Como Funciona na Prática
O engenheiro de I.A da IBM pede uma VM Bare Metal imunda e pelada. Precisa desenhar a malha via Terraform (VPCs, IP Elastic, Gateway, Security Group Portas) e se frustra instalando tudo. Por prequiça a equipe o impulsiona para o Amazon Sagemaker (PaaS). Presos no Vendor Lock-In do SageMaker o faturamento dispara U$ 1,800 no mês por instâncias rodando um notebook.

## 4. O Ganho Crolab (Injeção Zero Lock-In)
O que faremos? Destruíremos a premissa fundamental do Sagemaker.
O Binário *Go Node Crolab* vai embarcar perfeitamente dentro da AWS EC2 crua. Nosso desenvolvedor só locará a instância super rápida EC2 e conectará na Camada B de roteamento do CLI Crolab usando o Interceptador RPC. Ele programará o modelo ML a bilhões do conforto do seu Macbook sem jamais precisar abrir as telas burocráticas horríveis de provisionamento avançado dentro da Amazon Console.
