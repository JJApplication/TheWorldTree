package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Github   GithubConfig   `toml:"github"`
	Database DatabaseConfig `toml:"database"`
	Log      LogConfig      `toml:"log"`
}

type ServerConfig struct {
	HTTP struct {
		Host   string `toml:"host"`
		Port   int    `toml:"port"`
		Enable bool   `toml:"enable"`
	} `toml:"http"`
	GRPC struct {
		Address string `toml:"address"`
		Enable  bool   `toml:"enable"`
	}
}

type GithubConfig struct {
	Repositories []string `toml:"repositories"`
	Token        string   `toml:"token"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type LogConfig struct {
	Level string `toml:"level"`
}

var GlobalConfig *Config

func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if _, err := toml.Decode(string(data), &config); err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	GlobalConfig = &config
	return nil
}

func GetConfig() *Config {
	return GlobalConfig
}
