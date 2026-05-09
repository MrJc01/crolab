package sandbox

import (
	"fmt"
	"os/exec"
)

// NetworkJail cria e gerencia Network Namespaces isolados para bloquear a saída do kernel
type NetworkJail struct {
	Namespace string
	TapName   string
	Bridge    string
}

// NewNetworkJail instancia a configuração isolada.
func NewNetworkJail(id string) *NetworkJail {
	return &NetworkJail{
		Namespace: fmt.Sprintf("netns_%s", id),
		TapName:   fmt.Sprintf("tap_%s", id),
		Bridge:    "crolab_br0", // Ponte controlada que definiremos via iptables
	}
}

// Setup cria o Namespace, a interface TAP para o Firecracker e bloqueia saídas indevidas
func (n *NetworkJail) Setup() error {
	// 1. Cria o Namespace
	cmdNs := exec.Command("ip", "netns", "add", n.Namespace)
	if err := cmdNs.Run(); err != nil {
		return fmt.Errorf("falha ao criar netns: %v", err)
	}

	// 2. Cria Interface TAP vinculada ao Namespace
	cmdTap := exec.Command("ip", "netns", "exec", n.Namespace, "ip", "tuntap", "add", "dev", n.TapName, "mode", "tap")
	if err := cmdTap.Run(); err != nil {
		return fmt.Errorf("falha ao criar tap: %v", err)
	}

	// 3. Levanta a interface TAP
	cmdUp := exec.Command("ip", "netns", "exec", n.Namespace, "ip", "link", "set", n.TapName, "up")
	if err := cmdUp.Run(); err != nil {
		return fmt.Errorf("falha ao ativar tap: %v", err)
	}

	// OBS: A integração real com o Firecracker exigiria parear o Veth com a bridge do Host
	// e injetar regras rígidas de iptables. Ocultado aqui na abstração inicial para focar no contrato.

	return nil
}

// Teardown limpa toda a estrutura de rede do sandbox
func (n *NetworkJail) Teardown() error {
	cmd := exec.Command("ip", "netns", "del", n.Namespace)
	return cmd.Run()
}
