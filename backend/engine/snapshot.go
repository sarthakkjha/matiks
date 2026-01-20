// Package engine provides the ranking engine using a snapshot-based approach for O(1) lookups.
package engine

import (
	"sort"
	"sync"

	"matiks-leaderboard/cache"
)

type RankedEntry struct {
	UserID   string
	Username string
	Score    int
	Rank     int
}

type Snapshot struct {
	mu        sync.RWMutex
	entries   []RankedEntry
	rankIndex map[string]int
}

var Global = &Snapshot{
	entries:   make([]RankedEntry, 0),
	rankIndex: make(map[string]int),
}

func (s *Snapshot) Rebuild(data map[string]cache.Entry) {
	entries := make([]RankedEntry, 0, len(data))
	for id, e := range data {
		entries = append(entries, RankedEntry{
			UserID:   id,
			Username: e.Username,
			Score:    e.Score,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Username < entries[j].Username
		}
		return entries[i].Score > entries[j].Score
	})

	rankIndex := make(map[string]int, len(entries))
	currentRank := 1
	for i := range entries {
		if i > 0 && entries[i].Score != entries[i-1].Score {
			currentRank = i + 1
		}
		entries[i].Rank = currentRank
		rankIndex[entries[i].UserID] = currentRank
	}

	s.mu.Lock()
	s.entries = entries
	s.rankIndex = rankIndex
	s.mu.Unlock()
}

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

	result := make([]RankedEntry, end-start)
	copy(result, s.entries[start:end])
	return result, total
}

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

func (s *Snapshot) GetRank(userID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rankIndex[userID]
}

func (s *Snapshot) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
