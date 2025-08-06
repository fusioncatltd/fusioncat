package common

import "errors"

// Definitions of common errors used across the application
var (
	FusioncatErrRecordNotFound             = errors.New("DB record not found")
	FusioncatErrUniqueConstraintViolations = errors.New("Input violates unique constraint")
)
