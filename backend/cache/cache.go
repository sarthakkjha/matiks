// Package cache provides a thread-safe in-memory cache for user data.
package cache

import (
	"sort"
	"strings"
	"sync"
)

type Entry struct {
	Username string
	Score    int
}

type UserCache struct {
	mu   sync.RWMutex
	data map[string]Entry
}

var Global = &UserCache{
	data: make(map[string]Entry),
}

func (c *UserCache) Set(id string, entry Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[id] = entry
}

func (c *UserCache) Get(id string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.data[id]
	return e, ok
}

func (c *UserCache) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, id)
}

func (c *UserCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *UserCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]Entry)
}

type SearchResult struct {
	UserID   string
	Username string
	Score    int
}

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

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func (c *UserCache) GetAllWithIDs() map[string]Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]Entry, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

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
