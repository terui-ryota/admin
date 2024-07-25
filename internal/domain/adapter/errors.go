package adapter

import "errors"

var (
	ErrAdapterNotFound  = errors.New("adapter not found")
	ErrPermissionDenied = errors.New("permission denied")
)
