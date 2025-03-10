package config

import (
	"MydroX/project-v/pkg/config"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env     string   `yaml:"env"`
	Port    string   `yaml:"port"`
	DB      Database `yaml:"database"`
	JWT     JWT      `yaml:"jwt"`
	App     App      `yaml:"app"`
	Session Session  `yaml:"session"`
}

type App struct {
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
	AccessToken  AccessToken  `yaml:"access_token"`
	RefreshToken RefreshToken `yaml:"refresh_token"`
}

type AccessToken struct {
	Expiration int    `yaml:"expiration"`
	Secret     string `yaml:"secret"`
}

type RefreshToken struct {
	Expiration int    `yaml:"expiration"`
	Secret     string `yaml:"secret"`
}

type Session struct {
	Persistent bool `yaml:"persistent"`
}

func LoadConfig() (*Config, error) {
	f, err := config.Read()
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
