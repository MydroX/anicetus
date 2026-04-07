package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() *Config {
	return &Config{
		Port: "3000",
		Database: Database{
			Host:     "localhost",
			Port:     "5432",
			Username: "user",
			Password: "pass",
			Name:     "testdb",
		},
		Valkey: Valkey{
			Address: "localhost:6379",
		},
		Session: Session{
			Hash: Hash{
				Iterations:  3,
				Memory:      65536,
				Parallelism: 2,
				KeyLength:   32,
				SaltLength:  16,
			},
		},
	}
}

func TestConfig_Validate_Success(t *testing.T) {
	err := validConfig().Validate()
	require.NoError(t, err)
}

func TestConfig_Validate_MissingFields(t *testing.T) {
	tests := map[string]struct {
		modify      func(c *Config)
		expectedErr string
	}{
		"missing port": {
			modify:      func(c *Config) { c.Port = "" },
			expectedErr: "port is required",
		},
		"missing db host": {
			modify:      func(c *Config) { c.Database.Host = "" },
			expectedErr: "database host is required",
		},
		"missing db port": {
			modify:      func(c *Config) { c.Database.Port = "" },
			expectedErr: "database port is required",
		},
		"missing db username": {
			modify:      func(c *Config) { c.Database.Username = "" },
			expectedErr: "database username is required",
		},
		"missing db password": {
			modify:      func(c *Config) { c.Database.Password = "" },
			expectedErr: "database password is required",
		},
		"missing db name": {
			modify:      func(c *Config) { c.Database.Name = "" },
			expectedErr: "database name is required",
		},
		"missing valkey address": {
			modify:      func(c *Config) { c.Valkey.Address = "" },
			expectedErr: "valkey address is required",
		},
		"zero hash iterations": {
			modify:      func(c *Config) { c.Session.Hash.Iterations = 0 },
			expectedErr: "hash iterations must be greater than 0",
		},
		"zero hash memory": {
			modify:      func(c *Config) { c.Session.Hash.Memory = 0 },
			expectedErr: "hash memory must be greater than 0",
		},
		"zero hash parallelism": {
			modify:      func(c *Config) { c.Session.Hash.Parallelism = 0 },
			expectedErr: "hash parallelism must be greater than 0",
		},
		"zero hash key length": {
			modify:      func(c *Config) { c.Session.Hash.KeyLength = 0 },
			expectedErr: "hash key length must be greater than 0",
		},
		"zero hash salt length": {
			modify:      func(c *Config) { c.Session.Hash.SaltLength = 0 },
			expectedErr: "hash salt length must be greater than 0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig()
			tc.modify(cfg)
			err := cfg.Validate()
			require.Error(t, err)
			assert.Equal(t, tc.expectedErr, err.Error())
		})
	}
}
