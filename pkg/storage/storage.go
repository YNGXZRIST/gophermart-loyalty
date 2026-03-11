package storage

import "context"

type Storage[K comparable, V any] interface {
	Set(ctx context.Context, key K, value V) error
	Get(ctx context.Context, key K) (V, error)
	Delete(ctx context.Context, key K) error
	Has(ctx context.Context, key K) bool
	All(ctx context.Context) map[K]V
}
