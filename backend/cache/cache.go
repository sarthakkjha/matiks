// Package cache provides a thread-safe in-memory cache for user data.
// Uses sync.RWMutex for concurrent read/write access.
// This cache enables O(1) lookups and fast prefix searches without hitting the database.
package cache

import (
	"sort"
	"strings"
	"sync"
)

// Entry represents a cached user entry with username and score.
type Entry struct {
	Username string
	Score    int
}

// UserCache is a thread-safe in-memory cache backed by a map.
// DESIGN CHOICE:
// We use sync.RWMutex to allow multiple concurrent reads (Search/Get)
// while ensuring safe atomic writes (Update/Set).
//
// Performance:
// - Reads: O(1)
// - Writes: O(1)
// - Memory: ~1KB per user (10MB for 10K users) -> fits easily in RAM
type UserCache struct {
	mu   sync.RWMutex
	data map[string]Entry
}

// Global is the singleton cache instance used throughout the application.
// For millions of users, this would be replaced by a Redis client.
var Global = &UserCache{
	data: make(map[string]Entry),
}

// Set adds or updates a user in the cache.
// Thread-safe: acquires write lock.
func (c *UserCache) Set(id string, entry Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[id] = entry
}

// Get retrieves a user from the cache by ID.
// Returns (entry, true) if found, (Entry{}, false) otherwise.
// Thread-safe: acquires read lock.
func (c *UserCache) Get(id string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[id]
	return e, ok
}

// Delete removes a user from the cache.
// Thread-safe: acquires write lock.
func (c *UserCache) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, id)
}

// Size returns the number of entries in the cache.
// Thread-safe: acquires read lock.
func (c *UserCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Clear removes all entries from the cache.
// Thread-safe: acquires write lock.
func (c *UserCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]Entry)
}

// SearchResult contains user data returned from a search.
type SearchResult struct {
	UserID   string
	Username string
	Score    int
}

// SearchByPrefix performs a case-insensitive prefix search on usernames.
// Results are sorted by score descending and limited to 'limit' entries.
// Thread-safe: acquires read lock.
func (c *UserCache) SearchByPrefix(prefix string, limit int) []SearchResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	prefix = strings.ToLower(prefix)
	var results []SearchResult

	for id, e := range c.data {
		if strings.HasPrefix(strings.ToLower(e.Username), prefix) {
			results = append(results, SearchResult{
				UserID:   id,
				Username: e.Username,
				Score:    e.Score,
			})
		}
	}

	// Sort by score descending for relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

// GetAllWithIDs returns all entries with their IDs.
// Used by the snapshot manager for rebuilding rankings.
// Thread-safe: acquires read lock.
func (c *UserCache) GetAllWithIDs() map[string]Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]Entry, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// GetRandomIDs returns up to 'count' random user IDs.
// Used for bulk random updates.
func (c *UserCache) GetRandomIDs(count int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.data))
	for id := range c.data {
		ids = append(ids, id)
	}

	if count > len(ids) {
		count = len(ids)
	}
	return ids[:count]
}
