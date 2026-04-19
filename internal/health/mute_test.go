package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMute_DelegatesWhenUnmuted(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusUnhealthy}}
	mc := NewMuteChecker(inner, DefaultMuteConfig(), nil)
	r := mc.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %v", r.Status)
	}
}

func TestMute_ReturnsHealthyWhenMuted(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusUnhealthy}}
	mc := NewMuteChecker(inner, MuteConfig{Duration: 5 * time.Second}, nil)
	mc.Mute()
	r := mc.Check(context.Background(), "svc")
	if r.Status != StatusHealthy {
		t.Fatalf("expected healthy while muted, got %v", r.Status)
	}
}

func TestMute_UnmuteRestoresDelegation(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusUnhealthy}}
	mc := NewMuteChecker(inner, MuteConfig{Duration: 5 * time.Second}, nil)
	mc.Mute()
	mc.Unmute()
	r := mc.Check(context.Background(), "svc")
	if r.Status != StatusUnhealthy {
		t.Fatalf("expected unhealthy after unmute, got %v", r.Status)
	}
}

func TestMute_DefaultConfigOnZero(t *testing.T) {
	mc := NewMuteChecker(&fakeChecker{}, MuteConfig{}, nil)
	if mc.cfg.Duration != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", mc.cfg.Duration)
	}
}

func TestMuteHandler_ReturnsJSON(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	mc := NewMuteChecker(inner, DefaultMuteConfig(), nil)
	h := NewMuteStatusHandler(mc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	var resp muteStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Muted {
		t.Fatal("expected not muted")
	}
}

func TestMuteHandler_MuteAction(t *testing.T) {
	inner := &fakeChecker{result: Result{Status: StatusHealthy}}
	mc := NewMuteChecker(inner, DefaultMuteConfig(), nil)
	h := NewMuteStatusHandler(mc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/?action=mute", nil))
	if !mc.IsMuted() {
		t.Fatal("expected checker to be muted")
	}
}

func TestMuteHandler_UnknownAction(t *testing.T) {
	mc := NewMuteChecker(&fakeChecker{}, DefaultMuteConfig(), nil)
	h := NewMuteStatusHandler(mc, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/?action=noop", strings.NewReader("")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
