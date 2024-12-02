package config

import (
	"MydroX/project-v/pkg/config"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env  string   `yaml:"env"`
	Port string   `yaml:"port"`
	DB   Database `yaml:"database"`
	JWT  JWT      `yaml:"jwt"`
	App  App      `yaml:"app"`
}

type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Domain  string `yaml:"domain"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type JWT struct {
	ExpirationTime int    `yaml:"expiration_time"`
	Secret         string `yaml:"secret"`
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
