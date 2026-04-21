// Package health — trend checker
//
// TrendChecker wraps any Checker and annotates each Result with a "trend"
// metadata field that describes whether the health of a service is improving,
// degrading, or stable over a configurable rolling window of recent outcomes.
//
// # How it works
//
// After every check the outcome (healthy / unhealthy) is appended to an
// in-memory ring of up to Window entries per service.  Once at least MinSample
// entries have been collected the window is split in half and the healthy-rate
// of the second half is compared to the first:
//
//   - improving  — second-half rate exceeds first-half rate by > 10 %
//   - degrading  — second-half rate is below first-half rate by > 10 %
//   - stable     — rates are within 10 % of each other
//   - unknown    — not enough data yet
//
// The trend label is stored under the "trend" key in Result.Meta and is also
// surfaced via NewTrendStatusHandler.
package health
