package health

import (
	"errors"
	"testing"
	"time"
)

func TestStatus_String(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{StatusHealthy, "healthy"},
		{StatusUnhealthy, "unhealthy"},
		{StatusUnknown, "unknown"},
		{Status(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestResult_Healthy(t *testing.T) {
	now := time.Now()

	healthy := Result{Service: "svc", Status: StatusHealthy, CheckedAt: now}
	if !healthy.Healthy() {
		t.Error("expected Healthy() = true for StatusHealthy")
	}

	unhealthy := Result{Service: "svc", Status: StatusUnhealthy, CheckedAt: now}
	if unhealthy.Healthy() {
		t.Error("expected Healthy() = false for StatusUnhealthy")
	}

	unknown := Result{Service: "svc", Status: StatusUnknown, Err: errors.New("dial error"), CheckedAt: now}
	if unknown.Healthy() {
		t.Error("expected Healthy() = false for StatusUnknown")
	}
}

func TestResult_ErrPreserved(t *testing.T) {
	sentinel := errors.New("connection refused")
	r := Result{
		Service:   "",
		Status:    StatusUnknown,
		Err:       sentinel,
		CheckedAt: time.Now(),
	}
	if !errors.Is(r.Err, sentinel) {
		t.Errorf("expected Err to wrap sentinel, got %v", r.Err)
	}
}
