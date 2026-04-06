package cache

import "github.com/dgraph-io/ristretto/v2"

const (
	bufferItems = 64        // Default buffer items for Ristretto cache
	numCounters = 100000    // Number of counters for tracking key frequency
	maxCost     = 100 << 20 // Maximum cost for the cache, set to 100MB
)

func New() (*ristretto.Cache[string, string], error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		// For general usage, track frequency of 100,000 keys
		// This is sufficient for most applications but can be increased if needed
		NumCounters: numCounters,

		// Set maximum memory to 100MB instead of 1GB
		// More reasonable for general usage
		MaxCost: maxCost,

		// Default value is fine for most use cases
		BufferItems: bufferItems,

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

	return cache, nil
}
