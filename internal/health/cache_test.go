package health

import (
	"testing"
	"time"
)

func TestCache_MissOnEmpty(t *testing.T) {
	c := NewCache(5 * time.Second)
	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss on empty cache")
	}
}

func TestCache_HitAfterSet(t *testing.T) {
	c := NewCache(5 * time.Second)
	c.Set("svc", true)

	result, ok := c.Get("svc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if !result.Healthy {
		t.Error("expected healthy=true")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := NewCache(10 * time.Millisecond)
	c.Set("svc", true)

	time.Sleep(20 * time.Millisecond)

	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := NewCache(5 * time.Second)
	c.Set("svc", false)
	c.Invalidate("svc")

	_, ok := c.Get("svc")
	if ok {
		t.Fatal("expected cache miss after invalidation")
	}
}

func TestCache_MultipleServices(t *testing.T) {
	c := NewCache(5 * time.Second)
	c.Set("alpha", true)
	c.Set("beta", false)

	a, ok := c.Get("alpha")
	if !ok || !a.Healthy {
		t.Error("expected alpha to be healthy")
	}

	b, ok := c.Get("beta")
	if !ok || b.Healthy {
		t.Error("expected beta to be unhealthy")
	}
}

func TestCache_OverwriteEntry(t *testing.T) {
	c := NewCache(5 * time.Second)
	c.Set("svc", true)
	c.Set("svc", false)

	result, ok := c.Get("svc")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if result.Healthy {
		t.Error("expected overwritten value healthy=false")
	}
}
