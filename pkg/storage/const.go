package storage

import "errors"

var (
	// ErrNotFound is thrown when the key does not exist
	ErrNotFound = errors.New("not found")
)
