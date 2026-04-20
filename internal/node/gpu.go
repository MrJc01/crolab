// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package node

import (
	"fmt"
	"os/exec"
	"strings"
)

// GPUInfo represents a detected GPU on the host.
type GPUInfo struct {
	Index   string
	Name    string
	Memory  string
	Driver  string
}

// DetectGPUs uses nvidia-smi to list available GPUs.
// Returns empty slice if no NVIDIA GPUs or nvidia-smi not installed.
func DetectGPUs() []GPUInfo {
	out, err := exec.Command("nvidia-smi",
		"--query-gpu=index,gpu_name,memory.total,driver_version",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, ", ", 4)
		if len(parts) < 4 {
			continue
		}
		gpus = append(gpus, GPUInfo{
			Index:  strings.TrimSpace(parts[0]),
			Name:   strings.TrimSpace(parts[1]),
			Memory: strings.TrimSpace(parts[2]) + " MiB",
			Driver: strings.TrimSpace(parts[3]),
		})
	}
	return gpus
}

// DockerGPUFlags returns Docker flags to restrict a container to specific GPUs.
// If gpuIndex is empty, gives access to all GPUs.
func DockerGPUFlags(gpuIndex string) []string {
	if gpuIndex == "" {
		return []string{"--gpus", "all"}
	}
	return []string{"--gpus", fmt.Sprintf("\"device=%s\"", gpuIndex)}
}
