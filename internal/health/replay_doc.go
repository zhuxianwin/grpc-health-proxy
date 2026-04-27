// Package health — replay checker.
//
// ReplayChecker wraps any Checker and keeps a bounded, time-windowed log of
// every result it has produced. The log can be inspected via History() or
// exposed over HTTP with NewReplayStatusHandler.
//
// Typical use-cases:
//   - Debugging intermittent failures by inspecting recent health history.
//   - Feeding historical data into trend or slope checkers.
//   - Audit trails for compliance.
//
// Configuration:
//
//	Window   – how far back results are retained (default 30 s).
//	Capacity – maximum results stored per service (default 100).
package health
