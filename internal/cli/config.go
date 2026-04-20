// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Name     string `mapstructure:"name"`
	Address  string `mapstructure:"address"`
	Token    string `mapstructure:"token"`
	Provider string `mapstructure:"provider"`
	Priority int    `mapstructure:"priority"`
}

type CrolabConfig struct {
	CloudToken    string         `mapstructure:"cloud_token"`
	CloudAPI      string         `mapstructure:"cloud_api"`
	DefaultServer string         `mapstructure:"default_server"`
	Servers       []ServerConfig `mapstructure:"servers"`
}

var cfgFile string

func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfgFolder := filepath.Join(home, ".crolab")
	os.MkdirAll(cfgFolder, 0755)
	cfgFile = filepath.Join(cfgFolder, "config.yaml")

	viper.SetConfigFile(cfgFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			viper.Set("cloud_token", "")
			viper.Set("cloud_api", "")
			viper.Set("default_server", "")
			viper.Set("servers", []ServerConfig{})
			viper.WriteConfigAs(cfgFile)
		}
	}
}

// SetConfigFile overrides the config path (used by tests).
func SetConfigFile(path string) {
	cfgFile = path
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.ReadInConfig()
}

func LoadConfig() (CrolabConfig, error) {
	var config CrolabConfig
	err := viper.Unmarshal(&config)
	return config, err
}

func SaveConfig(config CrolabConfig) error {
	viper.Set("cloud_token", config.CloudToken)
	viper.Set("cloud_api", config.CloudAPI)
	viper.Set("default_server", config.DefaultServer)
	viper.Set("servers", config.Servers)
	return viper.WriteConfigAs(cfgFile)
}

func AddServer(name, address, token, provider string, priority int) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	for i, s := range cfg.Servers {
		if s.Name == name {
			cfg.Servers[i].Address = address
			cfg.Servers[i].Token = token
			cfg.Servers[i].Provider = provider
			cfg.Servers[i].Priority = priority
			return SaveConfig(cfg)
		}
	}

	cfg.Servers = append(cfg.Servers, ServerConfig{
		Name:     name,
		Address:  address,
		Token:    token,
		Provider: provider,
		Priority: priority,
	})

	if cfg.DefaultServer == "" || priority == 1 {
		cfg.DefaultServer = name
	}

	return SaveConfig(cfg)
}

func RemoveServer(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	newServers := []ServerConfig{}
	removed := false
	for _, s := range cfg.Servers {
		if s.Name != name {
			newServers = append(newServers, s)
		} else {
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("servidor '%s' não encontrado", name)
	}

	cfg.Servers = newServers

	if cfg.DefaultServer == name {
		if len(newServers) > 0 {
			cfg.DefaultServer = newServers[0].Name
		} else {
			cfg.DefaultServer = ""
		}
	}

	return SaveConfig(cfg)
}

func SetDefault(name string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	found := false
	for _, s := range cfg.Servers {
		if s.Name == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("servidor '%s' não encontrado", name)
	}
	cfg.DefaultServer = name
	return SaveConfig(cfg)
}

// ListServers returns all servers sorted by priority (lowest first).
func ListServers() ([]ServerConfig, string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, "", err
	}
	sorted := make([]ServerConfig, len(cfg.Servers))
	copy(sorted, cfg.Servers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	return sorted, cfg.DefaultServer, nil
}

// GetServer returns a specific server by name.
func GetServer(name string) (*ServerConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = cfg.DefaultServer
	}

	if name == "" {
		return nil, fmt.Errorf("nenhum servidor configurado. Use: crolab config add <nome> <ip:porta> <token>")
	}

	for _, s := range cfg.Servers {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("servidor '%s' não encontrado. Veja: crolab config ls", name)
}

// GetBestServer returns servers sorted by priority for failover.
func GetBestServer() ([]ServerConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	if len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("nenhum servidor configurado. Use: crolab config add")
	}
	sorted := make([]ServerConfig, len(cfg.Servers))
	copy(sorted, cfg.Servers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	return sorted, nil
}

// GenerateCrolabHash creates a cryptographic token for P2P auth.
func GenerateCrolabHash() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cl_" + hex.EncodeToString(b), nil
}
