package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/crolab/core/internal/cli"
	"github.com/spf13/viper"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.yaml")
	viper.Reset()
	viper.SetConfigFile(cfgPath)
	viper.SetConfigType("yaml")
	viper.Set("servers", []cli.ServerConfig{})
	viper.Set("default_server", "")
	viper.Set("cloud_token", "")
	viper.WriteConfigAs(cfgPath)
	cli.SetConfigFile(cfgPath)
}

func TestAddServer(t *testing.T) {
	setupTestConfig(t)
	err := cli.AddServer("node-a", "10.0.0.1:4422", "tok-a", "local", 1)
	if err != nil {
		t.Fatalf("AddServer falhou: %v", err)
	}
	servers, _, _ := cli.ListServers()
	if len(servers) != 1 {
		t.Fatalf("esperava 1 server, obteve %d", len(servers))
	}
	if servers[0].Name != "node-a" {
		t.Errorf("nome errado: %s", servers[0].Name)
	}
}

func TestAddServerUpdatesExisting(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("node-a", "10.0.0.1:4422", "tok", "local", 1)
	cli.AddServer("node-a", "10.0.0.2:4422", "tok2", "vastai", 2)

	servers, _, _ := cli.ListServers()
	if len(servers) != 1 {
		t.Fatalf("esperava 1 (update, não duplicata), obteve %d", len(servers))
	}
	if servers[0].Address != "10.0.0.2:4422" {
		t.Errorf("endereço não atualizou: %s", servers[0].Address)
	}
}

func TestRemoveServer(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("node-a", "1.1.1.1:4422", "", "local", 1)
	cli.AddServer("node-b", "2.2.2.2:4422", "", "local", 2)

	err := cli.RemoveServer("node-a")
	if err != nil {
		t.Fatalf("RemoveServer falhou: %v", err)
	}

	servers, _, _ := cli.ListServers()
	if len(servers) != 1 {
		t.Fatalf("esperava 1, obteve %d", len(servers))
	}
	if servers[0].Name != "node-b" {
		t.Errorf("server errado restante: %s", servers[0].Name)
	}
}

func TestRemoveNonExistent(t *testing.T) {
	setupTestConfig(t)
	err := cli.RemoveServer("fantasma")
	if err == nil {
		t.Error("deveria retornar erro ao remover server inexistente")
	}
}

func TestSetDefault(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("node-a", "1.1.1.1:4422", "", "local", 1)
	cli.AddServer("node-b", "2.2.2.2:4422", "", "local", 2)

	err := cli.SetDefault("node-b")
	if err != nil {
		t.Fatalf("SetDefault falhou: %v", err)
	}

	_, def, _ := cli.ListServers()
	if def != "node-b" {
		t.Errorf("default deveria ser node-b, é %s", def)
	}
}

func TestSetDefaultNonExistent(t *testing.T) {
	setupTestConfig(t)
	err := cli.SetDefault("fantasma")
	if err == nil {
		t.Error("deveria falhar com server inexistente")
	}
}

func TestListServersSortedByPriority(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("low", "1.1.1.1:4422", "", "local", 10)
	cli.AddServer("high", "2.2.2.2:4422", "", "vastai", 1)
	cli.AddServer("mid", "3.3.3.3:4422", "", "runpod", 5)

	servers, _, _ := cli.ListServers()
	if len(servers) != 3 {
		t.Fatalf("esperava 3, obteve %d", len(servers))
	}
	if servers[0].Name != "high" || servers[1].Name != "mid" || servers[2].Name != "low" {
		t.Errorf("ordem errada: %s, %s, %s", servers[0].Name, servers[1].Name, servers[2].Name)
	}
}

func TestGetServerByName(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("gpu-1", "10.0.0.5:4422", "secret", "aws", 1)

	s, err := cli.GetServer("gpu-1")
	if err != nil {
		t.Fatalf("GetServer falhou: %v", err)
	}
	if s.Token != "secret" || s.Provider != "aws" {
		t.Errorf("dados errados: token=%s provider=%s", s.Token, s.Provider)
	}
}

func TestGetServerDefault(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("default-node", "1.1.1.1:4422", "", "local", 1)

	s, err := cli.GetServer("") // empty = default
	if err != nil {
		t.Fatalf("GetServer('') falhou: %v", err)
	}
	if s.Name != "default-node" {
		t.Errorf("expected default-node, got %s", s.Name)
	}
}

func TestGetBestServerEmpty(t *testing.T) {
	setupTestConfig(t)
	_, err := cli.GetBestServer()
	if err == nil {
		t.Error("deveria falhar com config vazia")
	}
}

func TestGenerateCrolabHash(t *testing.T) {
	h1, err := cli.GenerateCrolabHash()
	if err != nil {
		t.Fatalf("GenerateCrolabHash falhou: %v", err)
	}
	h2, _ := cli.GenerateCrolabHash()

	if h1 == h2 {
		t.Error("dois hashes consecutivos são iguais — falha de entropia")
	}
	if len(h1) < 30 {
		t.Errorf("hash muito curto: %s", h1)
	}
	if h1[:3] != "cl_" {
		t.Errorf("hash deveria começar com cl_: %s", h1)
	}
}

func TestRemoveDefaultRotatesCorrectly(t *testing.T) {
	setupTestConfig(t)
	cli.AddServer("primary", "1.1.1.1:4422", "", "local", 1)
	cli.AddServer("backup", "2.2.2.2:4422", "", "local", 2)

	cli.RemoveServer("primary")
	_, def, _ := cli.ListServers()
	if def == "primary" {
		t.Error("default ainda aponta para server removido")
	}
}

func TestConfigPersistence(t *testing.T) {
	tmp := t.TempDir()
	cfgFile := filepath.Join(tmp, "persist-test.yaml")
	viper.Reset()
	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")
	viper.Set("servers", []cli.ServerConfig{})
	viper.Set("default_server", "")
	viper.Set("cloud_token", "")
	viper.WriteConfigAs(cfgFile)
	cli.SetConfigFile(cfgFile)

	cli.AddServer("persist-node", "9.9.9.9:4422", "tok", "gcp", 3)

	// Verify file exists and has content
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		t.Fatalf("config file não existe: %v", err)
	}
	if len(data) < 10 {
		t.Error("config file está vazio ou muito pequeno")
	}
}
