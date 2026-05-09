# 🖥️ Requisitos de Infraestrutura (Bare Metal vs Cloud)

Para rodar a Sandbox do Crolab (Fase 4), você **não pode usar servidores comuns que já são virtualizados sem suporte a KVM**.

## O Que é KVM e Por Que Ele Importa?
KVM (Kernel-based Virtual Machine) é o módulo do Linux que permite ao sistema agir como um hipervisor de hardware.
Como o Crolab utiliza **Firecracker MicroVMs** para garantir a segurança máxima contra malwares do usuário, nós dependemos fisicamente da presença do dispositivo `/dev/kvm`.

### ❌ O que NÃO funciona (Máquinas Incompatíveis):
- Instâncias EC2 padrões da AWS de baixo custo (família T2/T3).
- Droplets padrão da DigitalOcean.
- Contêineres Docker (Rodar o Crolab puramente em um Mac usando Docker Desktop fará com que a criação da máquina do Firecracker falhe, pois contêineres não emulam o hardware do hipervisor com segurança sem overhead insano).

### ✅ O que FUNCIONA (Máquinas Ideais para Produção):
- **Servidores Bare Metal:** Máquinas físicas reais alugadas na Hetzner, OVH, Equinix, etc.
- **AWS Instâncias "Metal":** Como as instâncias `i3.metal`, `c5n.metal` (Essas expõem o hardware diretamente).
- **Provedores que ativam Virtualização Aninhada (Nested Virtualization):** Alguns VPS premium no Google Cloud (GCP) permitem habilitar a tag `--enable-nested-virtualization` durante a criação da instância.

## Configuração do Sistema Operacional (OS)
Recomendamos o **Ubuntu 22.04 LTS** ou **24.04 LTS**.

### 1. Preparação Crítica do Kernel Host
Para que os módulos do Crolab (`internal/sandbox`) não quebrem, seu servidor físico precisa ter os seguintes recursos habilitados no Kernel do Linux:

```bash
# 1. Instale as ferramentas de KVM e Cgroups
sudo apt update
sudo apt install -y qemu-kvm libvirt-daemon-system firecracker bridge-utils cgroup-tools iptables

# 2. Atribua o usuário que rodará o Crolab ao grupo kvm
sudo usermod -aG kvm $USER

# 3. Garanta que as permissões do socket estão corretas (para ler sem root absoluto quando possível)
sudo chown root:kvm /dev/kvm
sudo chmod 660 /dev/kvm
```

### 2. Validando se a Máquina está Pronta
Rode este comando no terminal do seu provedor. A saída TEM que ser maior que 0.
```bash
egrep -c '(vmx|svm)' /proc/cpuinfo
```
*(Se for 0, seu servidor não suporta KVM virtualizado nativamente. Aborte a operação e contrate um Bare Metal).*

---
**Próximo Passo:** Com o hardware validado, veja como compilar o projeto e embutir o frontend em [02-BUILD-EMPACOTAMENTO.md](02-BUILD-EMPACOTAMENTO.md).
