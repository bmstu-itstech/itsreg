package shared

import "errors"

var (
	ErrPermissionDenied       = errors.New("permission denied")
	ErrIllegalStateTransition = errors.New("illegal state transition")
)
