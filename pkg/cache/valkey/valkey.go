package valkey

import (
	"github.com/valkey-io/valkey-go"
)

func NewClient(addr string) (valkey.Client, error) {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
