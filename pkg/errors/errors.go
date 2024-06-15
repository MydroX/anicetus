package errors

import "fmt"

var (
	// ErrNotFound is the error for entity not found
	ErrNotFound = fmt.Errorf("not found")
)
