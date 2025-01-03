package cerrors

import "errors"

// sentinel errors.
var (
	ErrValueRequired = errors.New("required")
	ErrInvalid       = errors.New("invalid")
)