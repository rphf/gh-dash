package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"charm.land/log/v2"
)

// HiddenPRStore keeps the hidden PRs in a state file: the updatedAt of each
// PR when it was hidden, by URL. A hidden PR comes back when it changes.
type HiddenPRStore struct {
	mu       sync.RWMutex
	saveMu   sync.Mutex
	saving   sync.WaitGroup
	entries  map[string]time.Time
	filePath string
}

func newHiddenPRStore(filePath string) *HiddenPRStore {
	s := &HiddenPRStore{entries: map[string]time.Time{}, filePath: filePath}
	if filePath == "" {
		return s
	}
	b, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error("Failed to load hidden PRs", "err", err)
		}
		return s
	}
	if err := json.Unmarshal(b, &s.entries); err != nil {
		log.Error("Failed to parse hidden PRs", "err", err)
		s.entries = map[string]time.Time{}
	}
	return s
}

// Hide hides the PR at url until it changes after updatedAt.
func (s *HiddenPRStore) Hide(url string, updatedAt time.Time) {
	s.mu.Lock()
	s.entries[url] = updatedAt
	s.mu.Unlock()
	s.saving.Go(s.save)
}

// Unhide shows the PR at url again.
func (s *HiddenPRStore) Unhide(url string) {
	s.mu.Lock()
	delete(s.entries, url)
	s.mu.Unlock()
	s.saving.Go(s.save)
}

// IsHidden tells if the PR at url is hidden. A PR whose updatedAt is after
// the time it was hidden is not hidden anymore, and its entry is removed.
func (s *HiddenPRStore) IsHidden(url string, updatedAt time.Time) bool {
	s.mu.RLock()
	hiddenAt, ok := s.entries[url]
	s.mu.RUnlock()
	if !ok || !updatedAt.After(hiddenAt) {
		return ok
	}
	s.mu.Lock()
	if cur, ok := s.entries[url]; ok && cur.Equal(hiddenAt) {
		delete(s.entries, url)
	}
	s.mu.Unlock()
	s.saving.Go(s.save)
	return false
}

func (s *HiddenPRStore) save() {
	if s.filePath == "" {
		return
	}
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	s.mu.RLock()
	b, err := json.MarshalIndent(s.entries, "", "  ")
	s.mu.RUnlock()
	if err == nil {
		err = os.MkdirAll(filepath.Dir(s.filePath), 0o755)
	}
	if err == nil {
		tmp := s.filePath + ".tmp"
		if err = os.WriteFile(tmp, b, 0o644); err == nil {
			err = os.Rename(tmp, s.filePath)
		}
	}
	if err != nil {
		log.Error("Failed to save hidden PRs", "err", err)
	}
}

var (
	hiddenPRStore     *HiddenPRStore
	hiddenPRStoreOnce sync.Once
)

// GetHiddenPRStore returns the hidden PRs store, kept in
// $XDG_STATE_HOME/gh-dash/hidden-prs.json.
func GetHiddenPRStore() *HiddenPRStore {
	hiddenPRStoreOnce.Do(func() {
		path, err := getStateFilePath("hidden-prs.json")
		if err != nil {
			log.Error("Failed to get state file path for hidden PRs", "err", err)
		}
		hiddenPRStore = newHiddenPRStore(path)
	})
	return hiddenPRStore
}

// SetHiddenPRStoreForTesting replaces the hidden PRs store and returns a
// function that restores it.
func SetHiddenPRStoreForTesting(filePath string) func() {
	GetHiddenPRStore()
	old := hiddenPRStore
	hiddenPRStore = newHiddenPRStore(filePath)
	return func() { hiddenPRStore = old }
}
