package gate

import "errors"

// Sentinel errors returned by Gate.Authorize.
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrNoPolicyDefined = errors.New("no policy defined for resource")
)
