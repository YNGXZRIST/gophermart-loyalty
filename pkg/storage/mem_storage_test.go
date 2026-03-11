package storage

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestMemStorage_Delete(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name    string
		s       *MemStorage[K, V]
		key     K
		wantErr error
	}
	tests := []testCase[string, int]{
		{
			name: "isset key",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:     "test_counter",
			wantErr: nil,
		},
		{
			name: "not isset key",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:     "test_counter_2",
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Delete(context.Background(), tt.key); !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_Get(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name    string
		s       *MemStorage[K, V]
		key     K
		want    V
		wantErr error
	}
	tests := []testCase[string, int]{
		{
			name: "isset value",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:     "test_counter",
			want:    100,
			wantErr: nil,
		},
		{
			name: "not isset value",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:     "test_counter_2",
			want:    0,
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Get(context.Background(), tt.key)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_Has(t *testing.T) {

	type testCase[K string, V int] struct {
		name string
		s    *MemStorage[K, V]
		key  K
		want bool
	}
	tests := []testCase[string, int]{
		{
			name: "isset key",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:  "test_counter",
			want: true,
		},
		{
			name: "not isset key",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			key:  "test_counter_2",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Has(context.Background(), tt.key); got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		key   K
		value V
	}
	type testCase[K comparable, V any] struct {
		name    string
		s       *MemStorage[K, V]
		args    args[K, V]
		wantErr bool
		wantVal V
	}
	tests := []testCase[string, int]{
		{
			name: "set isset key",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			args: struct {
				key   string
				value int
			}{key: "test_counter", value: 200},
			wantErr: false,
			wantVal: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.Set(context.Background(), tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			val, _ := tt.s.Get(context.Background(), tt.args.key)
			if !reflect.DeepEqual(val, tt.wantVal) {
				t.Errorf("Get() got = %v, want %v", val, tt.wantVal)
			}

		})
	}
}

func TestNewMemStorage(t *testing.T) {
	t.Run("new mem storage", func(t *testing.T) {
		got := NewMemStorage[string, int]()
		if got == nil {
			t.Fatal("NewMemStorage() returned nil")
		}
		if got.data == nil {
			t.Fatal("NewMemStorage() data map is nil")
		}
		if len(got.data) != 0 {
			t.Errorf("NewMemStorage() data should be empty, got %d elements", len(got.data))
		}
	})
}

func TestMemStorage_All(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name string
		s    *MemStorage[K, V]
		want map[K]V
	}
	tests := []testCase[string, int]{
		{
			name: "all",
			s: &MemStorage[string, int]{
				data: map[string]int{
					"test_counter": 100,
				},
			},
			want: map[string]int{
				"test_counter": 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.All(context.Background()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() = %v, want %v", got, tt.want)
			}
		})
	}
}
