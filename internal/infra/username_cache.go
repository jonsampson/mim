package infra

import (
	"context"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// UsernameCache provides efficient caching of PID to username mappings
// to avoid expensive user.LookupId() calls, especially in enterprise
// environments with network-based authentication (LDAP/AD).
type UsernameCache struct {
	cache   map[uint32]string
	mutex   sync.RWMutex
	timeout time.Duration
}

// NewUsernameCache creates a new username cache with reasonable defaults
func NewUsernameCache() *UsernameCache {
	return &UsernameCache{
		cache:   make(map[uint32]string),
		timeout: 100 * time.Millisecond, // Reasonable timeout for network calls
	}
}

// GetUsername retrieves the username for a given PID, using cache when possible
// and falling back to expensive lookups with timeout protection.
func (uc *UsernameCache) GetUsername(pid uint32) string {
	// Check cache first (fast path)
	uc.mutex.RLock()
	if username, exists := uc.cache[pid]; exists {
		uc.mutex.RUnlock()
		return username
	}
	uc.mutex.RUnlock()

	// Get UID for this PID (relatively fast - single /proc read)
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return "?"
	}

	uids, err := proc.Uids()
	if err != nil || len(uids) == 0 {
		return "?"
	}

	uid := uids[0] // Use real UID

	// Check if we've already cached this UID for another PID
	uc.mutex.RLock()
	for cachedPid, username := range uc.cache {
		if cachedPid != pid {
			// Check if this cached PID has the same UID
			if cachedProc, err := process.NewProcess(int32(cachedPid)); err == nil {
				if cachedUids, err := cachedProc.Uids(); err == nil && len(cachedUids) > 0 && cachedUids[0] == uid {
					// Same UID found - reuse the username
					uc.mutex.RUnlock()
					uc.mutex.Lock()
					uc.cache[pid] = username // Cache for this PID too
					uc.mutex.Unlock()
					return username
				}
			}
		}
	}
	uc.mutex.RUnlock()

	// Expensive lookup with timeout protection
	username := uc.lookupUsernameWithTimeout(uid)

	// Cache the result
	uc.mutex.Lock()
	uc.cache[pid] = username
	uc.mutex.Unlock()

	return username
}

// lookupUsernameWithTimeout performs the expensive user.LookupId call with timeout protection
func (uc *UsernameCache) lookupUsernameWithTimeout(uid uint32) string {
	ctx, cancel := context.WithTimeout(context.Background(), uc.timeout)
	defer cancel()

	done := make(chan string, 1)
	go func() {
		if u, err := user.LookupId(strconv.Itoa(int(uid))); err == nil {
			done <- u.Username
		} else {
			done <- strconv.Itoa(int(uid)) // Fallback to UID string
		}
	}()

	select {
	case username := <-done:
		return username
	case <-ctx.Done():
		// Timeout - return UID as fallback
		return strconv.Itoa(int(uid))
	}
}

// Clear removes old entries from the cache to prevent unlimited growth
func (uc *UsernameCache) Clear() {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()
	
	// Simple strategy: clear everything if cache gets too large
	if len(uc.cache) > 1000 {
		uc.cache = make(map[uint32]string)
	}
}

// Size returns the current number of cached entries (for monitoring/debugging)
func (uc *UsernameCache) Size() int {
	uc.mutex.RLock()
	defer uc.mutex.RUnlock()
	return len(uc.cache)
}