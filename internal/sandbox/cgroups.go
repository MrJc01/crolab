package sandbox

import (
	"fmt"
	"os/exec"
)

// CgroupBuilder ajuda a construir jails Cgroups V2 para limitar processos
type CgroupBuilder struct {
	Name      string
	CPUQuota  int64 // Em microsegundos (ex: 50000 = 50% de 1 núcleo)
	MemoryMax int64 // Em bytes
	PidsMax   int64 // Máximo de PIDs no grupo
}

// NewCgroupBuilder cria uma nova estrutura para um cgroup
func NewCgroupBuilder(name string) *CgroupBuilder {
	return &CgroupBuilder{
		Name:      name,
		CPUQuota:  100000,   // 100% de 1 vCPU default
		MemoryMax: 536870912, // 512 MB default
		PidsMax:   100,      // Prevenção primária contra fork bombs
	}
}

// Apply aplica as restrições no host (requer privilégios e cgroups v2 montado em /sys/fs/cgroup)
func (c *CgroupBuilder) Apply() error {
	path := fmt.Sprintf("/sys/fs/cgroup/%s", c.Name)
	
	// Cria o diretório do Cgroup
	cmd := exec.Command("mkdir", "-p", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha ao criar cgroup: %v", err)
	}

	// Limita Memória
	cmdMem := exec.Command("sh", "-c", fmt.Sprintf("echo %d > %s/memory.max", c.MemoryMax, path))
	if err := cmdMem.Run(); err != nil {
		return fmt.Errorf("falha ao limitar memory.max: %v", err)
	}

	// Limita CPU (formato: max <period> -> ex: 100000 100000)
	cmdCPU := exec.Command("sh", "-c", fmt.Sprintf("echo '%d 100000' > %s/cpu.max", c.CPUQuota, path))
	if err := cmdCPU.Run(); err != nil {
		return fmt.Errorf("falha ao limitar cpu.max: %v", err)
	}

	// Limita PIDs
	cmdPids := exec.Command("sh", "-c", fmt.Sprintf("echo %d > %s/pids.max", c.PidsMax, path))
	if err := cmdPids.Run(); err != nil {
		return fmt.Errorf("falha ao limitar pids.max: %v", err)
	}

	return nil
}

// AddPID adiciona um processo específico para dentro do Cgroup restrito
func (c *CgroupBuilder) AddPID(pid int) error {
	path := fmt.Sprintf("/sys/fs/cgroup/%s/cgroup.procs", c.Name)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %d > %s", pid, path))
	return cmd.Run()
}
