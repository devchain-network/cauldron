package cerrors

import "errors"

// sentinel errors.
var (
	ErrValueRequired = errors.New("value required")
	ErrInvalid       = errors.New("value is invalid")
)
