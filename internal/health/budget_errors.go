package health

import "errors"

// ErrBudgetExhausted is returned when the error budget for a service is spent.
var ErrBudgetExhausted = errors.New("error budget exhausted")
