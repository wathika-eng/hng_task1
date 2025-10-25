package database

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

var (
	// ErrExists returned when a value already exists
	ErrExists = errors.New("value already exists")
	// ErrNotFound returned when value not found
	ErrNotFound = errors.New("not found")
)

// StoreInterface defines storage operations used by the API
type StoreInterface interface {
	Save(v *Value) error
	GetByValue(value string) (*Value, bool)
	GetByHash(hash string) (*Value, bool)
	GetAll() []*Value
	DeleteByValue(value string) error
}

// memory-backed store (existing behavior)
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

// Gorm-backed store for Postgres persistence
type GormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (g *GormStore) Save(v *Value) error {
	// check by id or value
	var existing Value
	if err := g.db.Where("id = ? OR value = ?", v.ID, v.Value).First(&existing).Error; err == nil {
		return ErrExists
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return g.db.Create(v).Error
}

func (g *GormStore) GetByValue(value string) (*Value, bool) {
	var v Value
	if err := g.db.Where("value = ?", value).First(&v).Error; err != nil {
		return nil, false
	}
	return &v, true
}

func (g *GormStore) GetByHash(hash string) (*Value, bool) {
	var v Value
	if err := g.db.First(&v, "id = ?", hash).Error; err != nil {
		return nil, false
	}
	return &v, true
}

func (g *GormStore) GetAll() []*Value {
	var vals []Value
	if err := g.db.Find(&vals).Error; err != nil {
		return nil
	}
	out := make([]*Value, 0, len(vals))
	for i := range vals {
		out = append(out, &vals[i])
	}
	return out
}

func (g *GormStore) DeleteByValue(value string) error {
	res := g.db.Where("value = ?", value).Delete(&Value{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
