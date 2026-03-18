package storage

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Storage[K comparable, V any] interface {
	Set(ctx context.Context, key K, value V) error
	Get(ctx context.Context, key K) (V, error)
	Delete(ctx context.Context, key K) error
	Has(ctx context.Context, key K) bool
	All(ctx context.Context) map[K]V
}
type MemStorage[K comparable, V any] struct {
	data map[K]V
	mu   sync.Mutex
}

func NewMemStorage[K comparable, V any]() *MemStorage[K, V] {
	return &MemStorage[K, V]{data: make(map[K]V), mu: sync.Mutex{}}
}

func (s *MemStorage[K, V]) Set(ctx context.Context, key K, value V) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

func (s *MemStorage[K, V]) Get(ctx context.Context, key K) (V, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if !ok {
		var zero V
		return zero, ErrNotFound
	}
	return v, nil
}
func (s *MemStorage[K, V]) All(ctx context.Context) map[K]V {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := make(map[K]V, len(s.data))
	for k, v := range s.data {
		r[k] = v
	}
	return r
}

func (s *MemStorage[K, V]) Delete(ctx context.Context, key K) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return ErrNotFound
	}
	delete(s.data, key)
	return nil
}

func (s *MemStorage[K, V]) Has(ctx context.Context, key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
	return ok
}
