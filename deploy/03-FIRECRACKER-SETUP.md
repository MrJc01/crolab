# 🛡️ Configuração do Sandbox Firecracker

Saber como iniciar a microVM é a parte mais crítica do Crolab. Como o Crolab promete segurança isolada, precisamos preparar a infraestrutura base do sistema operacional hospedeiro.

O Firecracker precisa de 3 coisas para conseguir realizar o *boot* de uma máquina nova em menos de 50ms:
1. O executável VMM do Firecracker.
2. Um Kernel do Linux cru (`vmlinux`).
3. Uma imagem base do Sistema de Arquivos (RootFS / `rootfs.ext4`) contendo as linguagens que os usuários vão usar (Python, Node).

## 1. Baixando o Firecracker

Faça o download do binário estático no seu Bare Metal host:
```bash
wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.7.0/firecracker-v1.7.0-x86_64.tgz
tar -xvf firecracker-v1.7.0-x86_64.tgz
sudo mv release-v1.7.0-x86_64/firecracker-v1.7.0-x86_64 /usr/local/bin/firecracker
```

## 2. Kernel "Cru" Otimizado (vmlinux)
O Firecracker não faz boot com o kernel normal das distros porque ele foca no isolamento ultrarrápido (bypassando BIOS, PCI, etc).
Você precisa compilar ou baixar um Kernel não-comprimido (`vmlinux`) focado em MicroVMs.

```bash
# Baixando um Kernel pronto provido pelo Firecracker para testes
wget https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin
```

## 3. Preparando o RootFS (O Disco Base)
O usuário precisa de Python e das bibliotecas de Machine Learning (Pandas, Numpy, etc). Você deve criar um sistema de arquivos ext4 que contém o Alpine ou Ubuntu Minified.

```bash
# Baixando um rootfs pré-construído (apenas exemplo)
wget https://s3.amazonaws.com/spec.ccfc.min/img/hello/fsfiles/hello-rootfs.ext4
```
**Importante:** A sua imagem real de `rootfs` precisa ter um script no boot (`init`) que inicia a comunicação do **ZeroMQ** atrelada a ela!

## 4. Configurando a Ponte de Rede (Network Jail)
Como vimos no `internal/sandbox/network.go`, o Crolab amarra a máquina à uma bridge `crolab_br0`. Ela precisa existir fisicamente.

```bash
# 1. Cria a ponte principal do Crolab
sudo ip link add crolab_br0 type bridge
sudo ip link set crolab_br0 up

# 2. Permita que a bridge encaminhe pacotes usando NAT de saída para a internet
sudo iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
sudo iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
sudo iptables -A FORWARD -i crolab_br0 -o eth0 -j ACCEPT
```
*Atenção: A interface `eth0` acima deve ser trocada pelo nome real da interface de internet do seu Bare Metal (ex: `ens33`)*.

---
**Próximo Passo:** Se a VM for desligada, perdemos os arquivos se eles não forem persistidos. O próximo passo é instalar o MinIO (S3) para recuperar os notebooks via nuvem: [04-MINIO-S3-SETUP.md](04-MINIO-S3-SETUP.md).
