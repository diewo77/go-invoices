package gate

import (
	"context"
	"sync"
	"time"
)

// CachedResolver wraps a ProfileResolver with TTL-based caching.
// This avoids hitting the database on every authorization check.
type CachedResolver[U comparable] struct {
	inner ProfileResolver[U]
	cache map[U]*cacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type cacheEntry struct {
	profile   Profile
	expiresAt time.Time
}

// NewCachedResolver wraps a resolver with caching.
// ttl is how long profiles are cached before re-fetching.
func NewCachedResolver[U comparable](inner ProfileResolver[U], ttl time.Duration) *CachedResolver[U] {
	return &CachedResolver[U]{
		inner: inner,
		cache: make(map[U]*cacheEntry),
		ttl:   ttl,
	}
}

// Resolve returns the profile for the given user, using cache if available.
func (r *CachedResolver[U]) Resolve(ctx context.Context, user U) (Profile, error) {
	// Check cache first (read lock)
	r.mu.RLock()
	entry, ok := r.cache[user]
	r.mu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.profile, nil
	}

	// Cache miss or expired - fetch from inner resolver
	profile, err := r.inner.Resolve(ctx, user)
	if err != nil {
		return nil, err
	}

	// Store in cache (write lock)
	r.mu.Lock()
	r.cache[user] = &cacheEntry{
		profile:   profile,
		expiresAt: time.Now().Add(r.ttl),
	}
	r.mu.Unlock()

	return profile, nil
}

// Invalidate removes a user from the cache.
// Call this when a user's profile assignment changes.
func (r *CachedResolver[U]) Invalidate(user U) {
	r.mu.Lock()
	delete(r.cache, user)
	r.mu.Unlock()
}

// InvalidateAll clears the entire cache.
// Call this when profile permissions are modified.
func (r *CachedResolver[U]) InvalidateAll() {
	r.mu.Lock()
	r.cache = make(map[U]*cacheEntry)
	r.mu.Unlock()
}
