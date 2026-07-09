package runtime

import "errors"

var (
	ErrRuntimeNotConfigured = errors.New("discord runtime is not configured")
	ErrInvalidRuntimeEvent  = errors.New("discord runtime event is invalid")
	ErrGatewaySmokeNotReady = errors.New("gateway smoke test did not observe ready")
)
