package customerrors

import (
	"errors"
)

// Server Error Messages
var (
	ErrInternalServer = errors.New("internal server error")

	//Client Error Messages
	ErrBadRequest = errors.New("bad request")

	ErrInvalidId = errors.New("invalid UUID")
)
