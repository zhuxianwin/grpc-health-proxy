package health

import "errors"

// ErrAdmissionLimitExceeded is returned when the maximum number of concurrent
// health checks has been reached and the incoming check is rejected.
var ErrAdmissionLimitExceeded = errors.New("admission limit exceeded: too many concurrent health checks")
