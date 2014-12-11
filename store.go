package main

import (
	"sync"
)

// The Store interface defines methods to manipulate items.
type Store interface {
	Get(id uint64) *Series
	GetAll() []*Series
	Add(p *Series) (uint64, error)
	Update(p *Series) error
}

// Thread-safe in-memory map.
type SeriesStore struct {
	sync.RWMutex
	m map[uint64]*Series
}

// GetAll returns all Seriess from memory
func (store *SeriesStore) GetAll() []*Series {
	store.RLock()
	defer store.RUnlock()
	if len(store.m) == 0 {
		return nil
	}
	ar := make([]*Series, len(store.m))
	i := 0
	for _, v := range store.m {
		ar[i] = v
		i++
	}
	return ar
}

// Get returns a single Series identified by its id, or nil.
func (store *SeriesStore) Get(id uint64) *Series {
	store.RLock()
	defer store.RUnlock()
	return store.m[id]
}

// Add stores a new Series and returns its newly generated id, or an error.
func (store *SeriesStore) Add(p *Series) (uint64, error) {
	store.Lock()
	defer store.Unlock()
	// Store it
	store.m[p.ID] = p
	return p.ID, nil
}

// Update updates Series and returns nil
func (store *SeriesStore) Update(p *Series) error {
	store.Lock()
	defer store.Unlock()
	store.m[p.ID] = p
	return nil
}

func (store *SeriesStore) Delete(id uint64) {
	store.Lock()
	defer store.Unlock()
	delete(store.m, id)
}
