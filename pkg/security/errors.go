package security

import "errors"

var (
	// ErrNoStorage indicates no storage backend is available
	ErrNoStorage = errors.New("no storage backend available")
	
	// ErrNoPrimary indicates no primary backend is configured
	ErrNoPrimary = errors.New("no primary backend configured")
	
	// ErrNotFound indicates a secret was not found
	ErrNotFound = errors.New("secret not found")
	
	// ErrReadOnly indicates the backend is read-only
	ErrReadOnly = errors.New("storage backend is read-only")
	
	// ErrNotAvailable indicates the backend is not available
	ErrNotAvailable = errors.New("storage backend not available")
)