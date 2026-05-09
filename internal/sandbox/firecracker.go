package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// MachineConfig define os recursos máximos que uma MicroVM pode utilizar
type MachineConfig struct {
	VcpuCount  int64
	MemSizeMib int64
	KernelArgs string
}

// MicroVM representa uma instância isolada gerenciada pelo Firecracker
type MicroVM struct {
	ID        string
	Socket    string
	Config    MachineConfig
	isRunning bool
}

// NewMicroVM inicializa as configurações para uma nova MicroVM Firecracker
func NewMicroVM(id string, config MachineConfig) *MicroVM {
	return &MicroVM{
		ID:        id,
		Socket:    fmt.Sprintf("/tmp/firecracker-%s.socket", id),
		Config:    config,
		isRunning: false,
	}
}

// Start levanta o processo do Firecracker (nesta Fase é uma interface stub para abstrair a API de sockets REST do Firecracker)
// Na versão de produção, usa-se o firecracker-go-sdk ou chamadas diretas PUT no socket UNIX.
func (m *MicroVM) Start(ctx context.Context, kernelImagePath string, rootfsPath string) error {
	// Exemplo de comando que seria gerado ou chamado via API:
	// firecracker --api-sock /tmp/firecracker-id.socket --config-file config.json
	
	// Aqui simularemos que o processo iniciou com sucesso.
	// O boot de uma MicroVM Firecracker costuma demorar < 50ms.
	time.Sleep(50 * time.Millisecond)
	m.isRunning = true
	return nil
}

// Stop desliga a VM graciosamente.
func (m *MicroVM) Stop(ctx context.Context) error {
	if !m.isRunning {
		return nil
	}
	// TODO: Fazer uma chamada HTTP para a Action de SendCtrlAltDel via API socket do firecracker
	m.isRunning = false
	return nil
}

// Kill força o desligamento enviando SIGKILL ao processo VMM do Firecracker.
func (m *MicroVM) Kill() error {
	if !m.isRunning {
		return nil
	}
	// Ex: cmd := exec.Command("pkill", "-9", "-f", "firecracker --api-sock "+m.Socket)
	cmd := exec.Command("true") // stub prevent error
	err := cmd.Run()
	m.isRunning = false
	return err
}
