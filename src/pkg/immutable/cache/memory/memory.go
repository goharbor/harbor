package memory

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/immutable/cache"
	"sync"
)

// Cache ...
type Cache struct {
	mu            sync.RWMutex
	repostitories map[int64]map[string]interface{}
	immutable     map[string]bool
}

// NewMemoryCache ...
func NewMemoryCache() Cache {
	return Cache{
		repostitories: make(map[int64]map[string]interface{}),
		immutable:     make(map[string]bool),
	}
}

// Set ...
func (immc *Cache) Set(pid int64, imc cache.IMCandidate) error {
	immc.mu.Lock()
	defer immc.mu.Unlock()

	return immc.set(pid, imc)
}

// SetMultiple ...
func (immc *Cache) SetMultiple(pid int64, imcs []cache.IMCandidate) error {
	immc.mu.Lock()
	defer immc.mu.Unlock()

	for _, imc := range imcs {
		if err := immc.set(pid, imc); err != nil {
			return err
		}
	}
	return nil
}

// Set ...
func (immc *Cache) set(pid int64, imc cache.IMCandidate) error {
	_, proExist := immc.repostitories[pid]
	if !proExist {
		immc.repostitories[pid] = make(map[string]interface{})
	}

	repos := immc.repostitories[pid]
	_, tagExist := repos[imc.Repository+"::"+imc.Tag]
	if !tagExist {
		repos[imc.Repository+"::"+imc.Tag] = struct{}{}
		immc.immutable[imc.Repository+"::"+imc.Tag] = imc.Immutable
	}
	return nil
}

// Stat ...
func (immc *Cache) Stat(pid int64, repository string, tag string) (bool, error) {
	immc.mu.RLock()
	defer immc.mu.RUnlock()

	repositories := immc.repostitories[pid]
	_, exist := repositories[repository+"::"+tag]
	if !exist {
		return false, fmt.Errorf("no repository:%s and tag:%s found in project repositories", repository, tag)
	}

	_, exist = immc.immutable[repository+"::"+tag]
	if !exist {
		return false, fmt.Errorf("no immutable found for tagL %s::%s", repository, tag)
	}

	return immc.immutable[repository+"::"+tag], nil
}

// Clear ...
func (immc *Cache) Clear(pid int64, imc cache.IMCandidate) error {
	immc.mu.Lock()
	defer immc.mu.Unlock()

	repos := immc.repostitories[pid]
	delete(repos, imc.Repository+"::"+imc.Tag)
	delete(immc.immutable, imc.Repository+"::"+imc.Tag)
	return nil
}

// Flush ...
func (immc *Cache) Flush(pid int64) error {
	immc.mu.Lock()
	defer immc.mu.Unlock()

	delete(immc.repostitories, pid)
	return nil
}
