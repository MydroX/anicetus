package config

import "errors"

func (c *Config) Validate() error {
	if c.Port == "" {
		return errors.New("port is required")
	}

	if err := c.Database.validate(); err != nil {
		return err
	}

	if err := c.Valkey.validate(); err != nil {
		return err
	}

	return c.Session.Hash.validate()
}

func (d *Database) validate() error {
	if d.Host == "" {
		return errors.New("database host is required")
	}

	if d.Port == "" {
		return errors.New("database port is required")
	}

	if d.Username == "" {
		return errors.New("database username is required")
	}

	if d.Password == "" {
		return errors.New("database password is required")
	}

	if d.Name == "" {
		return errors.New("database name is required")
	}

	return nil
}

func (v *Valkey) validate() error {
	if v.Address == "" {
		return errors.New("valkey address is required")
	}

	return nil
}

func (h *Hash) validate() error {
	if h.Iterations == 0 {
		return errors.New("hash iterations must be greater than 0")
	}

	if h.Memory == 0 {
		return errors.New("hash memory must be greater than 0")
	}

	if h.Parallelism == 0 {
		return errors.New("hash parallelism must be greater than 0")
	}

	if h.KeyLength == 0 {
		return errors.New("hash key length must be greater than 0")
	}

	if h.SaltLength == 0 {
		return errors.New("hash salt length must be greater than 0")
	}

	return nil
}
