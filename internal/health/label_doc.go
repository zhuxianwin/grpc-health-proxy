// Package health provides label enrichment via [NewLabelChecker].
//
// # Label Checker
//
// The label checker wraps any [Checker] and attaches a static set of
// key-value labels to every check invocation. Labels are logged at
// DEBUG level and can be inspected at runtime via [NewLabelStatusHandler].
//
// Labels are useful for annotating health-check results with deployment
// metadata such as environment, region, or team ownership without modifying
// the underlying checker.
//
// Example:
//
//	cfg := health.LabelConfig{
//		Labels: map[string]string{
//			"env":    "production",
//			"region": "us-east-1",
//		},
//	}
//	c := health.NewLabelChecker(inner, cfg, logger)
package health
