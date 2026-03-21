package server

import (
	"gophermart-loyalty/internal/gopherman/logger"
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")

	t.Run("success with env", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", ":8080")
		t.Setenv("DATABASE_URI", "postgres://localhost/gophermart")
		t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
		t.Setenv("ACCRUAL_WORKER_COUNT", "3")
		got, err := NewConfig([]string{})
		if err != nil {
			t.Errorf("NewConfig() error = %v", err)
			return
		}
		want := &Config{}
		want.Address = ":8080"
		want.DatabaseURL = "postgres://localhost/gophermart"
		want.AccrualAddress = "http://accrual:8080"
		want.Mode = logger.TypeModeDevelopment
		want.AccrualWorkerCount = 3
		if !reflect.DeepEqual(got, want) {
			t.Errorf("NewConfig() got = %+v, want %+v", got, want)
		}
	})

	t.Run("success with flags and env for accrual", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", "")
		t.Setenv("DATABASE_URI", "")
		t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
		got, err := NewConfig([]string{"-a", "0.0.0.0:8080", "-d", "postgres://user:pass@host/db"})
		if err != nil {
			t.Errorf("NewConfig() error = %v", err)
			return
		}
		if got.Address != "0.0.0.0:8080" || got.DatabaseURL != "postgres://user:pass@host/db" || got.AccrualAddress != "http://accrual:8080" {
			t.Errorf("NewConfig() got = %+v", got)
		}
	})

	t.Run("invalid flag", func(t *testing.T) {
		_, err := NewConfig([]string{"-unknown"})
		if err == nil {
			t.Error("NewConfig() expected error for unknown flag")
		}
	})

	t.Run("missing database URL", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", ":8080")
		t.Setenv("DATABASE_URI", "")
		t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
		_, err := NewConfig([]string{})
		if err == nil {
			t.Error("NewConfig() expected error when database URL is empty")
		}
	})

	t.Run("missing accrual address", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", ":8080")
		t.Setenv("DATABASE_URI", "postgres://localhost/db")
		t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")
		_, err := NewConfig([]string{})
		if err == nil {
			t.Error("NewConfig() expected error when accrual address is empty")
		}
	})
}
