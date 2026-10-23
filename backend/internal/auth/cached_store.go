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

// SessionLive is asked on every request; only live answers are kept, and
// EndSession drops that, so signing out takes effect on the next request.
func (c CachedStore) SessionLive(ctx context.Context, sid string) (bool, error) {
	var live bool
	if c.Cache.Get(ctx, cacheSessionKey+sid, &live) {
		return live, nil
	}
	live, err := c.Store.SessionLive(ctx, sid)
	if err == nil && live {
		c.Cache.Set(ctx, cacheSessionKey+sid, true)
	}
	return live, err
}

func (c CachedStore) EndSession(ctx context.Context, sid string) error {
	err := c.Store.EndSession(ctx, sid)
	c.Cache.Del(ctx, cacheSessionKey+sid)
	return err
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
