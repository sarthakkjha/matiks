// Package engine provides the ranking engine for the leaderboard.
// Uses a snapshot-based approach for O(1) rank lookups after O(N log N) rebuilds.
// The snapshot is immutable once built, allowing lock-free reads.
package engine

import (
	"sort"
	"sync"

	"matiks-leaderboard/cache"
)

// RankedEntry represents a user with their computed rank.
type RankedEntry struct {
	UserID   string
	Username string
	Score    int
	Rank     int
}

// Snapshot manages a pre-sorted leaderboard with O(1) rank lookups.
//
// ALGORITHM:
// 1. Full Sort: We sort all users by Score DESC (O(N log N)) background process.
// 2. Rank Index: We build a HashMap (UserID -> Rank) for O(1) lookups.
// 3. Atomic Swap: We swap the pointer to the new snapshot instantly.
//
// READ AVAILABILITY:
// - Reads during rebuild are served from the OLD snapshot (zero downtime).
// - No read locks are required on the entries slice once visible.
type Snapshot struct {
	mu        sync.RWMutex
	entries   []RankedEntry  // Sorted by score descending
	rankIndex map[string]int // userID -> rank (1-indexed)
}

// Global is the singleton snapshot instance.
var Global = &Snapshot{
	entries:   make([]RankedEntry, 0),
	rankIndex: make(map[string]int),
}

// Rebuild recreates the snapshot from cache data.
// Users with the same score receive the same rank.
// This is O(N log N) due to sorting but only runs on score changes.
func (s *Snapshot) Rebuild(data map[string]cache.Entry) {
	// Build entries list
	entries := make([]RankedEntry, 0, len(data))
	for id, e := range data {
		entries = append(entries, RankedEntry{
			UserID:   id,
			Username: e.Username,
			Score:    e.Score,
		})
	}

	// Sort by score descending, then username for stable ordering
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Username < entries[j].Username
		}
		return entries[i].Score > entries[j].Score
	})

	// Assign ranks (same score = same rank)
	rankIndex := make(map[string]int, len(entries))
	currentRank := 1
	for i := range entries {
		if i > 0 && entries[i].Score != entries[i-1].Score {
			currentRank = i + 1
		}
		entries[i].Rank = currentRank
		rankIndex[entries[i].UserID] = currentRank
	}

	// Atomic swap
	s.mu.Lock()
	s.entries = entries
	s.rankIndex = rankIndex
	s.mu.Unlock()
}

// GetLeaderboard returns paginated leaderboard entries.
// Thread-safe: acquires read lock.
func (s *Snapshot) GetLeaderboard(page, limit int) ([]RankedEntry, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.entries)
	start := (page - 1) * limit
	if start >= total {
		return []RankedEntry{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}

	// Return a copy to prevent data races
	result := make([]RankedEntry, end-start)
	copy(result, s.entries[start:end])
	return result, total
}

// GetTop returns the top N entries.
// Thread-safe: acquires read lock.
func (s *Snapshot) GetTop(n int) []RankedEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.entries) {
		n = len(s.entries)
	}
	result := make([]RankedEntry, n)
	copy(result, s.entries[:n])
	return result
}

// GetRank returns the rank for a user ID.
// Returns 0 if user not found.
// O(1) lookup from the rank index.
func (s *Snapshot) GetRank(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rankIndex[userID]
}

// Size returns the number of entries in the snapshot.
func (s *Snapshot) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
