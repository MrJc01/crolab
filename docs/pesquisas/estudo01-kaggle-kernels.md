# Estudo Competitivo: Kaggle Kernels

## 1. Arquitetura Base (Cache na Borda)
Kaggle, embora deva lealdade à Alphabet (Google), difere do Colab brutalmente na ingestão de dados. A arquitetura de Edge Storage deles prevê que PetaBytes de Datasets do Hub deles se mantenham em cache ao lado dos containers de Kernel.
O container não precisa baixar os dados; o Hypervisor apenas realiza um Symlink / Mount instantâneo dos 50GB do Dataset competitivo.

## 2. Média de Valor/Custo
Totalmente subsidiado pela extração de dados e CrowdSourcing empresarial (O Modelo Gratuito Limitado). Eles liberam Cotas semanais de GPU P100/T4 (ex: 30 a 40 horas). O verdadeiro preço é faturado no lado de Data-mining do Google coletando os Kernels públicos e nas marcas que patrocinam prêmios no site.

## 3. Como Funciona na Prática
Você edita um notebook similar ao Jupyter, focado primariamente em submissão em Lote. Diferente do Colab, o Kernel inteiro pode rodar "no escuro" enquanto seu browser fecha, executando "Save and Run All" em Background de graça com timeout de 12 horas.

## 4. A Visão Crolab (O Que Copiar / Insight)
A arquitetura de *Save and Run All em Background* é exatamente o pilar do Módulo Crolab Server `SubmitJob`. Mas o segredo é o **Mount de Volume Nativo**. A Crolab pode estabelecer uma política em que os Nodes hospedadores (O "The Tank") funcionem como Caches Bitorrent de Datasets famosos (ex: ImageNet). No vasto P2P, em vez de enviar os pesos e imagens pela internet, nossos Nodes já possuirão esses datasets espelhados da HuggingFace, submetendo o Treinamento sem estrangular a banda do desenvolvedor.
