package health

import "errors"

// ErrLatchTripped is returned when a latchChecker is in the tripped state
// and refuses to delegate to the inner checker.
var ErrLatchTripped = errors.New("health: latch tripped")
