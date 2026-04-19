package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Name    string `mapstructure:"name"`
	Address string `mapstructure:"address"`
	Token   string `mapstructure:"token"`
}

type CrolabConfig struct {
	DefaultServer string         `mapstructure:"default_server"`
	Servers       []ServerConfig `mapstructure:"servers"`
}

var cfgFile string

// InitConfig garante que os apontamentos P2P existam e gerencia state.
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
			viper.Set("default_server", "")
			viper.Set("servers", []ServerConfig{})
			viper.WriteConfigAs(cfgFile)
		}
	}
}

func LoadConfig() (CrolabConfig, error) {
	var config CrolabConfig
	err := viper.Unmarshal(&config)
	return config, err
}

func SaveConfig(config CrolabConfig) error {
	viper.Set("default_server", config.DefaultServer)
	viper.Set("servers", config.Servers)
	return viper.WriteConfigAs(cfgFile)
}

func AddServer(name, address, token string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	for i, s := range cfg.Servers {
		if s.Name == name {
			cfg.Servers[i].Address = address
			cfg.Servers[i].Token = token
			return SaveConfig(cfg)
		}
	}

	cfg.Servers = append(cfg.Servers, ServerConfig{
		Name:    name,
		Address: address,
		Token:   token,
	})

	if cfg.DefaultServer == "" {
		cfg.DefaultServer = name
	}

	return SaveConfig(cfg)
}

func GetServer(name string) (*ServerConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = cfg.DefaultServer
	}

	for _, s := range cfg.Servers {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("Servidor Ocioso ['%s'] não está homologado na malha config. Utilize crolab config add antes de atirar.", name)
}
