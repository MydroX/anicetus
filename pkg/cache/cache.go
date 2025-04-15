package cache

import "github.com/dgraph-io/ristretto/v2"

func New() (*ristretto.Cache[string, string], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		// For general usage, track frequency of 100,000 keys
		// This is sufficient for most applications but can be increased if needed
		NumCounters: 1e5,

		// Set maximum memory to 100MB instead of 1GB
		// More reasonable for general usage
		MaxCost: 100 << 20,

		// Default value is fine for most use cases
		BufferItems: 64,

		// Use cost calculation based on item size
		Cost: func(value string) int64 {
			return int64(len(value))
		},

		// Enable metrics collection
		Metrics: true,
	})
	if err != nil {
		return nil, err
	}
	defer cache.Close()

	return cache, nil
}
