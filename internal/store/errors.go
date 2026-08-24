package store

import "errors"

var (
	ErrNotFound   = errors.New("record not found")
	ErrConflict   = errors.New("record conflict")
	ErrState      = errors.New("invalid state transition")
	ErrValidation = errors.New("validation failed")
)
