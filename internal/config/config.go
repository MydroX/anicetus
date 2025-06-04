//revive:disable:max-public-structs
package config

type Config struct {
	Env      string   `yaml:"env"`
	Port     string   `yaml:"port"`
	App      App      `yaml:"app"`
	Database Database `yaml:"database"`
	JWT      JWT      `yaml:"jwt"`
	Session  Session  `yaml:"session"`
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
	Secret       string       `yaml:"secret"`
	SkewSeconds  int          `yaml:"skew"`
	Issuer       string       `yaml:"issuer"`
	AccessToken  AccessToken  `yaml:"access_token"`
	RefreshToken RefreshToken `yaml:"refresh_token"`
}

type AccessToken struct {
	Expiration int `yaml:"expiration"`
}

type RefreshToken struct {
	Expiration int `yaml:"expiration"`
}

type Session struct {
	Hash       Hash `yaml:"hash"`
	IP         IP   `yaml:"ip"`
	Persistent bool `yaml:"persistent"`
}

type Hash struct {
	SaltLength  uint32 `yaml:"salt_length"`
	Iterations  uint32 `yaml:"iterations"`
	Memory      uint32 `yaml:"memory"`
	Parallelism uint8  `yaml:"parallelism"`
	KeyLength   uint32 `yaml:"key_length"`
}

type IP struct {
	Salt string `yaml:"salt"`
}
