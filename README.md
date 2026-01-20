# Matiks Leaderboard System

A **high-performance leaderboard** designed to handle **10,000+ users** with **instant searches**, **high-throughput updates**, and **real-time rankings**.

---

## Requirements Fulfilled

| Requirement | Solution | Performance |
|-------------|----------|-------------|
| 10,000+ users | MongoDB + in-memory cache |  10,100 users seeded |
| Username search | Prefix search with caching |  O(N) with early termination |
| Rating updates | Debounced batch processing |  4,000+ updates/sec |
| Real time leaderboard | Snapshot based rankings |  O(1) rank lookups |

---

## Optimization #1: Username Search

**Challenge**: Search 10,000 usernames quickly without overwhelming the database.

**Solution**: In-Memory Cache with Prefix Matching

```text
User types: "sha"
     │
     ▼
┌────────────────────────────────────────┐
│          IN-MEMORY CACHE               │
│   HashMap: userID → {username, score}  │
│                                        │
│   For each entry:                      │
│     if username.toLowerCase()          │
│        .startsWith("sha")  ◀─── O(1)   │
│     then add to results                │
│                                        │
│   Sort by score DESC                   │
│   Return top 20                        │
└────────────────────────────────────────┘
     │
     ▼
Results: ShadowKing, ShadowPro, Shadow123...
```

**Why this works**:
- **Zero database queries** - all data in RAM
- **Case-insensitive** - `strings.ToLower()` comparison
- **Sorted by relevance** - highest-rated users first
- **Thread-safe** - `sync.RWMutex` allows concurrent reads

**Code**: [`cache/cache.go`](backend/cache/cache.go) → `SearchByPrefix()`

---

## Optimization #2: High-Throughput Writes

**Challenge**: Handle hundreds of rating updates per second without performance degradation.

**Solution**: Debounced Batch Rebuilding

```text
                    WITHOUT OPTIMIZATION
┌─────────────────────────────────────────────────┐
│  Update 1 → Rebuild snapshot → O(N log N)       │
│  Update 2 → Rebuild snapshot → O(N log N)       │
│  Update 3 → Rebuild snapshot → O(N log N)       │
│  ...                                            │
│  100 updates = 100 rebuilds = TOO SLOW          │
└─────────────────────────────────────────────────┘

                     WITH OPTIMIZATION  
┌─────────────────────────────────────────────────┐
│  Update 1 ─┐                                    │
│  Update 2 ─┼─▶ Cache updated instantly (O(1))   │
│  Update 3 ─┤        │                           │
│  ...       │        ▼                           │
│  Update N ─┘   Debounce Timer (100ms)           │
│                     │                           │
│                     ▼                           │
│              SINGLE REBUILD                     │
│              O(N log N) × 1 = FAST              │
└─────────────────────────────────────────────────┘
```


**Results**:
| Updates | Without Batching | With Batching |
|---------|------------------|---------------|
| 100 | 100 O(N log N) rebuilds | 1 rebuild |

**Measured Performance**: **4,385 updates/second** (tested with bulk update API)

**Code**: [`services/leaderboard.go`](backend/services/leaderboard.go) → `scheduleRebuild()`

---

## Optimization #3: Real-Time Leaderboard

**Challenge**: Show accurate rankings after updates without slow database queries.

**Solution**: Snapshot-Based Ranking Engine

```text
┌─────────────────────────────────────────────────────────────┐
│                    SNAPSHOT MANAGER                         │
│                                                             │
│  ┌──────────────────────┐    ┌─────────────────────────────┐│
│  │   Sorted Array       │    │      Rank Index             ││
│  │   (by score DESC)    │    │      (HashMap)              ││
│  │                      │    │                             ││
│  │  [0] Champion: 5000  │    │  userID → rank              ││
│  │  [1] Legend: 4999    │    │  "abc123" → 1               ││
│  │  [2] Master: 4998    │    │  "def456" → 2               ││
│  │  [3] User_A: 4997    │    │  "ghi789" → 3               ││
│  │  [4] User_B: 4997    │◀──▶│  ...                        ││
│  │  ...                 │    │                             ││
│  └──────────────────────┘    └─────────────────────────────┘│
│         │                              │                    │
│         ▼                              ▼                    │
│  GetLeaderboard(page, limit)    GetRank(userID)             │
│  O(1) slice operation           O(1) map lookup             │
└─────────────────────────────────────────────────────────────┘
```

**Key Features**:
1. **Tied Rankings**: Same score = Same rank (e.g., two users at #4)
2. **O(1) Rank Lookups**: Pre-computed rank index
3. **Non-Blocking Reads**: Snapshot swap is atomic
4. **Pagination**: Slice of pre-sorted array

**Code**: [`engine/snapshot.go`](backend/engine/snapshot.go) → `Rebuild()`, `GetRank()`

---

## 🏗️ Architecture

```text
┌────────────┐      ┌─────────────────────────────────────────────────┐
│  Frontend  │      │                  GO BACKEND                     │
│  (React    │─────▶│                                                 │
│  Native)   │      │   ┌─────────────────────────────────────────┐   │
└────────────┘      │   │             Gin Router                  │   │
                    │   │  /api/leaderboard, /api/users/search    │   │ 
                    │   └─────────────────┬───────────────────────┘   │
                    │                     │                           │
                    │   ┌─────────────────▼───────────────────────┐   │
                    │   │           Leaderboard Service           │   │
                    │   │  - Debounced rebuilds                   │   │
                    │   │  - CRUD operations                      │   │
                    │   └───────┬───────────────────┬─────────────┘   │
                    │           │                   │                 │
                    │   ┌───────▼──────┐    ┌──────▼──────┐           │
                    │   │   Cache      │    │  Snapshot   │           │
                    │   │(sync.RWMutex)│    │  Manager    │           │
                    │   │  O(1) r/w    │    │ O(1) ranks  │           │
                    │   └──────────────┘    └─────────────┘           │
                    │                                                 │
                    │   ┌──────────────────────────────────────────┐  │
                    │   │              MongoDB                     │  │
                    │   │         (Persistent Storage)             │  │
                    │   └──────────────────────────────────────────┘  │
                    └─────────────────────────────────────────────────┘
```

---

## 📈 Performance Demo

The Admin panel includes a **Mass Update** feature to demonstrate high-throughput handling:

1. Go to **Admin** tab
2. Enter number of users (e.g., 1000)
3. Click **Random Ratings** or **Set to Value**
4. See real-time performance: **updates/second**

---

## 🚀 Scaling Approach

**Current**: In-process cache (Go `sync.RWMutex`) for 10K users - fastest and simplest for demo.

**At scale**: The cache layer is modular and can be swapped to **Redis** (HSET for data, ZSET for rankings) without changing core logic. Same patterns, different infrastructure.

