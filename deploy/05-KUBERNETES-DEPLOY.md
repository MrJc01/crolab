# ☸️ Orquestração com Kubernetes (Deploy Final)

Na "Fase 6" do Crolab, construímos os Custom Resource Definitions (CRDs) do Kubernetes para que o provisionamento do Firecracker seja automático e resiliente através de múltiplos Servidores Bare Metal (Cluster).

Quando você passa de 1 servidor (Host) para 20 servidores físicos, instalar e gerenciar pontes de rede e IP de MicroVMs manualmente torna-se um pesadelo.

O Kubernetes resolve isso, transformando seus 20 servidores Bare Metal em um "Cérebro Global Único".

## 1. Aplicando o Crolab Kernel CRD no Cluster
Com um cluster K8s vivo (você pode usar o `K3s` no seu servidor para economizar memória), aplique a definição que criamos no projeto:

```bash
kubectl apply -f deploy/k8s/crolab-crd.yaml
```
Esse comando "ensina" ao Kubernetes uma nova palavra mágica: `CrolabKernel` ou `ckernel`.

## 2. Invocando Máquinas Dinamicamente
Agora, o Backend Go do Crolab, ao invés de rodar comandos de sistema (`os.exec`) via bash, pode se conectar à API do Kubernetes como um Pod autorizado e dizer: *"Ei K8s, crie uma máquina nova de 2GB para o usuário XYZ"*.

O JSON gerado pela API do Go vai injetar isso aqui no seu Cluster K8s:
```yaml
apiVersion: crolab.io/v1alpha1
kind: CrolabKernel
metadata:
  name: sessao-ai-julio-x1
  namespace: crolab-users
spec:
  machineType: "cpu-optimized"
  vcpus: 4
  memoryMb: 2048
  userID: "usr_abcd1234"
```

## 3. Monitoramento da Frota (Prometheus)

Outra imensa vantagem de rodar dentro do K8s é não perder as métricas de latência dos WebSockets e do ZeroMQ que desenvolvemos na Fase 6.

Já criamos as rotas em `internal/gateway/metrics.go` e o arquivo de alvo do *Prometheus* em `deploy/observability/prometheus.yml`.

### Como ativar:
1. Suba o pod do Prometheus oficial no Kubernetes e faça ele mapear para o nosso manifesto usando o comando normal do helm (ou suba via Docker compose apontando os caminhos corretos).
2. O Prometheus achará automaticamente todos os pods/instâncias marcadas com as labels do Crolab no cluster usando seu método `kubernetes_sd_configs`.
3. Você verá em tempo real no Grafana a quantidade de "Mensagens Encaminhadas" subindo vertiginosamente.

---
🚀 **Parabéns!** Você tem a visão completa do sistema Crolab rodando em alta escala (SRE Grade), protegida por Sandboxes MicroVM e orquestrada de forma elástica pelo Kubernetes.
