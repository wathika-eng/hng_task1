package database

import (
	"errors"
	"sync"
)

var (
	// ErrExists returned when a value already exists
	ErrExists = errors.New("value already exists")
	// ErrNotFound returned when value not found
	ErrNotFound = errors.New("not found")
)

type Store struct {
	mu      sync.RWMutex
	byHash  map[string]*Value
	byValue map[string]string // value -> hash
}

func NewStore() *Store {
	return &Store{
		byHash:  make(map[string]*Value),
		byValue: make(map[string]string),
	}
}

func (s *Store) Save(v *Value) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byHash[v.ID]; ok {
		return ErrExists
	}
	if _, ok := s.byValue[v.Value]; ok {
		return ErrExists
	}
	// store
	s.byHash[v.ID] = v
	s.byValue[v.Value] = v.ID
	return nil
}

func (s *Store) GetByValue(value string) (*Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.byValue[value]
	if !ok {
		return nil, false
	}
	v, ok := s.byHash[h]
	return v, ok
}

// GetByHash returns a Value by its SHA256 id
func (s *Store) GetByHash(hash string) (*Value, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.byHash[hash]
	return v, ok
}

func (s *Store) GetAll() []*Value {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*Value, 0, len(s.byHash))
	for _, v := range s.byHash {
		res = append(res, v)
	}
	return res
}

func (s *Store) DeleteByValue(value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.byValue[value]
	if !ok {
		return ErrNotFound
	}
	delete(s.byValue, value)
	delete(s.byHash, h)
	return nil
}
