package ristretto

import "github.com/dgraph-io/ristretto/v2"

const (
	bufferItems = 64
	numCounters = 100000
	maxCost     = 100 << 20 // 100MB
)

func NewClient() (*ristretto.Cache[string, string], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
		Cost: func(value string) int64 {
			return int64(len(value))
		},
	})
	if err != nil {
		return nil, err
	}

	return cache, nil
}
