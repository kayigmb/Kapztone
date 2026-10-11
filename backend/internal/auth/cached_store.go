package auth

import (
	"context"

	"github.com/kayigmb/chaintrack/backend/internal/cache"
)

// CachedStore keeps the two lookups made on every signed-in request in the
// cache. Only successful lookups are stored, so gaining access is always
// immediate; losing access is immediate as long as the write that removes it
// calls the matching Invalidate function. The cache's expiry is the backstop
// for changes made outside the API.
type CachedStore struct {
	Store
	Cache *cache.Cache
}

func scopeCacheKey(userID, slug string) string { return cacheScopePrefix + userID + ":" + slug }

func (c CachedStore) ActiveUser(ctx context.Context, id string) (User, error) {
	key := cacheUserPrefix + id
	var u User
	if c.Cache.Get(ctx, key, &u) {
		return u, nil
	}
	u, err := c.Store.ActiveUser(ctx, id)
	if err == nil {
		c.Cache.Set(ctx, key, u)
	}
	return u, err
}

func (c CachedStore) Scope(ctx context.Context, userID, slug string) (Scope, error) {
	key := scopeCacheKey(userID, slug)
	var s Scope
	if c.Cache.Get(ctx, key, &s) {
		return s, nil
	}
	s, err := c.Store.Scope(ctx, userID, slug)
	if err == nil {
		c.Cache.Set(ctx, key, s)
	}
	return s, err
}

// InvalidateUser must be called after any change to a user's account or to
// their scopes (status, deletion, role, membership).
func (c CachedStore) InvalidateUser(ctx context.Context, userID string) {
	c.Cache.Del(ctx, cacheUserPrefix+userID)
	c.Cache.DelPattern(ctx, scopeCacheKey(userID, "*"))
}

// InvalidateOrg must be called after any change to an organisation that a
// scope carries (status, name, deletion).
func (c CachedStore) InvalidateOrg(ctx context.Context, slug string) {
	c.Cache.DelPattern(ctx, scopeCacheKey("*", slug))
}

// InvalidateAllScopes must be called after roles or permissions change.
func (c CachedStore) InvalidateAllScopes(ctx context.Context) {
	c.Cache.DelPattern(ctx, cacheScopePrefix+"*")
}
