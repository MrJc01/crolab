# O Ecossistema de Base e Tecnologias Futuras

Este documento mapeia as ferramentas acopladas que serão essenciais na maturação e segurança do ecossistema e atuarão na Fase 2 da arquitetura.

## 1. O MVP: Docker Padrão 
Em vez de desenhar imagens altamente estritas de imediato ("Cromtainer"), o MVP permite que o usuário dispare scripts baseando-se em imagens abertas (ex: `nvidia/cuda:12.1.1-runtime-ubuntu22.04`).
A sincronização inicial de arquivos também seguirá um padrão nativo. O binário `crolab-cli` cria um simples '.zip' assinado do pacote e transfere via canal gRPC para a máquina destino na inicialização do job, revertendo a saída no final do processamento.

## 2. Visão de Futuro (Fase V2)

### A. Integração com Crompressor
O gargalo real de treinar LLMs ou Video Models não está em como o Docker empacota a imagem, está no peso astronômico de Subir (Egress) um `.pt` de milhões de parâmetros por ssh toda a vez que mexemos num código da Main. É na Fase V2 que o `Crompressor` substitui as transferências de Zip/Rsync por Sincronização de Delta baseada em Hashes preenchendo as lacunas sem tráfego redundante.

### B. Cromtainer
O amadurecimento dos Daemons e a necessidade de isolamento absoluto das permissões (para usuários alugando para terceiros) fará do Docker puro uma possível brecha caso os privilégios falhem. O conceito *Cromtainer* trará imagens cacheadas globalmente na rede restritamente configuráveis.

### C. Módulos / Plugins de Computação Free
Uma frente separada e inteiramente experimental focada em sequestros inofensivos de instâncias "free tier" de outras empresas (Acoplamento de Webhook com Jupyter ou API oculta de Colab). Como dependem de scrapers sujeitos à mudança ou cookies de navegadores voláteis, foram desentranhadas do "Core-Business" de venda de *Compute* e delegadas para testes futuros com arquitetura modular de Plugins.
