package config

type Config struct {
	Env     string   `yaml:"env"`
	Port    string   `yaml:"port"`
	DB      Database `yaml:"database"`
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

type Session struct {
	Persistent   bool         `yaml:"persistent"`
	AccessToken  AccessToken  `yaml:"access_token"`
	RefreshToken RefreshToken `yaml:"refresh_token"`
	HashConfig   HashConfig   `yaml:"hash_config"`
	IP           IP           `yaml:"ip"`
}

type HashConfig struct {
	SaltLength  int `yaml:"salt_length"`
	Iterations  int `yaml:"iterations"`
	Memory      int `yaml:"memory"`
	Parallelism int `yaml:"parallelism"`
	KeyLength   int `yaml:"key_length"`
}

type IP struct {
	Salt string `yaml:"salt"`
}

type AccessToken struct {
	Expiration int    `yaml:"expiration"`
	Secret     string `yaml:"secret"`
}

type RefreshToken struct {
	Expiration int    `yaml:"expiration"`
	Secret     string `yaml:"secret"`
}

// func LoadConfig() (*Config, error) {
// 	err := config.LoadConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// }
