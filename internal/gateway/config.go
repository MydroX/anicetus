package gateway

import (
	"MydroX/project-v/pkg/config"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env  string   `yaml:"env"`
	Port string   `yaml:"port"`
	DB   Database `yaml:"database"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func LoadConfig(serviceName string) (*Config, error) {
	f, err := config.Read(serviceName)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
