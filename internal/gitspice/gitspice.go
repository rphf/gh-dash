// Package gitspice reads stacks from git-spice in a local clone.
package gitspice

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"sync"
	"time"
)

const cacheTTL = 30 * time.Second

type branch struct {
	Name string `json:"name"`
	Down *struct {
		Name string `json:"name"`
	} `json:"down"`
	Change *struct {
		URL string `json:"url"`
	} `json:"change"`
}

type cacheEntry struct {
	parents map[string]string
	at      time.Time
}

var (
	mu    sync.Mutex
	cache = map[string]cacheEntry{}
)

// ParentURLs returns, for each PR tracked by git-spice in the repo at dir, the
// URL of the PR below it. It returns an empty map when git-spice is not
// installed or the repo does not use it.
func ParentURLs(dir string) map[string]string {
	mu.Lock()
	entry, ok := cache[dir]
	mu.Unlock()
	if ok && time.Since(entry.at) < cacheTTL {
		return entry.parents
	}

	parents := map[string]string{}
	if out, err := run(dir); err == nil {
		parents = parse(out)
	}

	mu.Lock()
	cache[dir] = cacheEntry{parents: parents, at: time.Now()}
	mu.Unlock()
	return parents
}

func run(dir string) ([]byte, error) {
	bin, err := exec.LookPath("git-spice")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		bin,
		"ls",
		"--all",
		"--json",
		"--no-prompt",
		"--no-cr-status",
		"--no-cr-comments",
	)
	cmd.Dir = dir
	return cmd.Output()
}

func parse(out []byte) map[string]string {
	urlByBranch := map[string]string{}
	var branches []branch
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		var b branch
		if err := json.Unmarshal(sc.Bytes(), &b); err != nil {
			continue
		}
		branches = append(branches, b)
		if b.Change != nil && b.Change.URL != "" {
			urlByBranch[b.Name] = b.Change.URL
		}
	}

	parents := map[string]string{}
	for _, b := range branches {
		if b.Change == nil || b.Down == nil {
			continue
		}
		if parentURL, ok := urlByBranch[b.Down.Name]; ok {
			parents[b.Change.URL] = parentURL
		}
	}
	return parents
}
