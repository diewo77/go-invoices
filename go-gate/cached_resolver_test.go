package gate_test

import (
	"context"
	"testing"
	"time"

	"github.com/diewo77/go-gate"
)

func TestCachedResolver_CachesProfile(t *testing.T) {
	inner := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "editor")
	inner.Set(1, profile)

	cached := gate.NewCachedResolver[uint](inner, 5*time.Minute)

	// First call - cache miss
	p1, err := cached.Resolve(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p1.Name() != "editor" {
		t.Errorf("expected 'editor', got '%s'", p1.Name())
	}

	// Modify inner resolver (simulate change)
	newProfile := gate.NewStaticProfile(1, "admin")
	inner.Set(1, newProfile)

	// Second call - should return cached value
	p2, err := cached.Resolve(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p2.Name() != "editor" {
		t.Errorf("expected cached 'editor', got '%s'", p2.Name())
	}
}

func TestCachedResolver_Invalidate(t *testing.T) {
	inner := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "editor")
	inner.Set(1, profile)

	cached := gate.NewCachedResolver[uint](inner, 5*time.Minute)

	// Populate cache
	_, _ = cached.Resolve(context.Background(), 1)

	// Modify inner
	newProfile := gate.NewStaticProfile(1, "admin")
	inner.Set(1, newProfile)

	// Invalidate cache for user 1
	cached.Invalidate(1)

	// Should now get new value
	p, err := cached.Resolve(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "admin" {
		t.Errorf("expected 'admin' after invalidation, got '%s'", p.Name())
	}
}

func TestCachedResolver_InvalidateAll(t *testing.T) {
	inner := gate.NewStaticResolver[uint]()
	inner.Set(1, gate.NewStaticProfile(1, "editor"))
	inner.Set(2, gate.NewStaticProfile(2, "viewer"))

	cached := gate.NewCachedResolver[uint](inner, 5*time.Minute)

	// Populate cache
	_, _ = cached.Resolve(context.Background(), 1)
	_, _ = cached.Resolve(context.Background(), 2)

	// Modify both
	inner.Set(1, gate.NewStaticProfile(1, "admin"))
	inner.Set(2, gate.NewStaticProfile(2, "admin"))

	// Invalidate all
	cached.InvalidateAll()

	// Both should return new values
	p1, _ := cached.Resolve(context.Background(), 1)
	p2, _ := cached.Resolve(context.Background(), 2)

	if p1.Name() != "admin" || p2.Name() != "admin" {
		t.Error("expected both profiles to be 'admin' after InvalidateAll")
	}
}

func TestCachedResolver_TTLExpiry(t *testing.T) {
	inner := gate.NewStaticResolver[uint]()
	inner.Set(1, gate.NewStaticProfile(1, "editor"))

	// Very short TTL
	cached := gate.NewCachedResolver[uint](inner, 10*time.Millisecond)

	// Populate cache
	_, _ = cached.Resolve(context.Background(), 1)

	// Modify inner
	inner.Set(1, gate.NewStaticProfile(1, "admin"))

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Should get new value
	p, _ := cached.Resolve(context.Background(), 1)
	if p.Name() != "admin" {
		t.Errorf("expected 'admin' after TTL expiry, got '%s'", p.Name())
	}
}
