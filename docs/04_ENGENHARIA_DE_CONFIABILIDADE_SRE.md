# Engenharia de Confiabilidade (SRE) e Diagnóstico Forense

Como a base fundamental do Crolab é criar abstrações orquestradas sobre provedores P2P onde instâncias e larguras de banda não possuem SLA corporativo garantido, a arquitetura deve prever falhas constantemente. Nossa diretriz SRE trata comportamento de provedores terceiros com desconfiança e resiliência intrínseca.

## Investigação SRE de Falha no Ecossistema Distribuído

Sempre que a malha de `crolab-nodes` apresentar comportamento inesperado, orfandade de jobs, desconexões ou OOM (Out-of-memory) no container da GPU, os diagnósticos são executados sistematicamente através das 4 camadas de abstração.

### Camada 1: Transporte/Rede (Malha de Túneis)
- **Diagnóstico Oculto**: Máquinas no Vast.ai operam sob CGNAT com portas assimétricas, acarretando em túneis WebSocket/gRPC quebrando (TCP timeouts e Keep-Alive failures).
- **Passos de Investigação**: Avaliação dos pacotes originados do `crolab-node` usando ping-pong de telemetria e rastreabilidade de handshake bidirecional (quem fecha a conexão?).
- **Solução (Desenho da Infra)**: O *Server* não disca para a instância (PULL). Em vez disso, o `crolab-node` abre túnel reverso e notifica o Server que está pronto para receber instruções via gRPC Streaming (PUSH reverso).
- **Prevenção SRE**: O Node aplica backoff exponencial para reconectar se o Crolab Server oscilar, garantindo persistência eterna.

### Camada 2: Aplicação (O Cérebro e Roteamento)
- **Diagnóstico Oculto**: Erros de autorização (401/403) invisíveis. Muitas vezes um "runner" P2P fica ocioso porque o Identity Hub barrou o token JWT após *timeout*, e o Node não está alertando a plataforma.
- **Passos de Investigação**: Rastrear Request-IDs do Identity Hub até a ingestão no Server. Validar a fila de *Jobs Pendentes* presa em memória.
- **Solução**: Validar *Heartbeats* com recertificação leve. Implementação rigorosa de logs contextuais `log.WithFields` focada no "Porquê" a requisição expirou.

### Camada 3: Persistência/Estado (Sincronização e Volumes)
- **Diagnóstico Oculto**: A instância GPU encerra abruptamente, o desenvolvedor perde o *Dataset*/Pesos de 12h de treino, acreditando que foi problema de aplicação, quando na verdade foi armazenamento efêmero do host descartado.
- **Passos de Investigação**: Checar a última assinatura delta (log transacional) do *Crompressor*, verificando a fragmentação do file-system da instância na checagem final do host morto.
- **Solução**: Uso de injeção de Daemons side-car de Backup contínuo em intervalos baseados no throughput (a cada 10 tokens / 5 minutos salva o log do checkpoint remotamente).
- **Prevenção SRE**: Instâncias efêmeras devem rodar sobre diretórios *tmpfs* limitados em tamanho ou monitorados rigidamente pelo Node Agent antes que saturem os discos limitados dos locadores do Vast.ai.

### Camada 4: Ambiente (O Crolab-Node e Cromtainer)
- **Diagnóstico Oculto**: Container GPU reporta `Unknown Timeout` ao dar Docker Run ou a GPU perde o acesso via `/dev/nvidia*`.
- **Passos de Investigação**: Ingressar com comando `docker events` remoto. Averiguar suporte passará ao driver via kernel do host alugado contra o driver exigido pelo Cromtainer (Mismatch).
- **Solução**: O script Bash de inicialização do nó precisa validar os binários, a API e as bibliotecas do sistema antes de sinalizar para o Server que "Estou Online e Seguro".
- **Prevenção SRE**: Um nó passa por teste de "Burn-in" de 10 segundos antes do orquestrador entregar a primeira carga paga do cliente, certificando saúde absoluta de Hardware.
