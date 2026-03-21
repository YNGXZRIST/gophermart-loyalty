package main

import (
	"context"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/logger"
	"testing"
	"time"

	"go.uber.org/zap"
)

func clearServerEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS",
		"MODE", "ACCRUAL_WORKER_COUNT", "ACCRUAL_USE_MOCK",
	} {
		t.Setenv(k, "")
	}
}

func TestInitConfig_validFlags(t *testing.T) {
	clearServerEnv(t)
	cfg, err := initConfig([]string{
		"-a", "127.0.0.1:8080",
		"-d", "postgres://user:pass@localhost:5432/dbname",
		"-r", "http://127.0.0.1:9000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Address != "127.0.0.1:8080" {
		t.Errorf("Address = %q", cfg.Address)
	}
	if cfg.DatabaseURL == "" || cfg.AccrualAddress == "" {
		t.Errorf("expected database and accrual URL set")
	}
}

func TestInitConfig_missingDatabaseURL(t *testing.T) {
	clearServerEnv(t)
	_, err := initConfig([]string{
		"-a", "127.0.0.1:8080",
		"-r", "http://127.0.0.1:9000",
	})
	if err == nil {
		t.Fatal("want error when DATABASE_URI / -d missing")
	}
}

func TestInitConfig_missingAccrualAddress(t *testing.T) {
	clearServerEnv(t)
	_, err := initConfig([]string{
		"-a", "127.0.0.1:8080",
		"-d", "postgres://localhost/db",
	})
	if err == nil {
		t.Fatal("want error when accrual address missing")
	}
}

func TestInitLogger_validMode(t *testing.T) {
	cfg := &server.Config{}
	cfg.Mode = logger.TypeModeTest
	lgr, err := initLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if lgr == nil {
		t.Fatal("nil logger")
	}
	_ = lgr.Sync()
}

func TestInitLogger_invalidMode(t *testing.T) {
	cfg := &server.Config{}
	cfg.Mode = "not-a-valid-mode"
	_, err := initLogger(cfg)
	if err == nil {
		t.Fatal("want error for invalid mode")
	}
}

func TestInitRepos(t *testing.T) {
	repos := initRepos(nil)
	if repos.User == nil || repos.Order == nil || repos.Withdrawal == nil {
		t.Fatal("all repositories must be non-nil")
	}
}

func TestInitHTTPHandler(t *testing.T) {
	repos := initRepos(nil)
	h := initHTTPHandler(nil, repos, zap.NewNop())
	if h == nil {
		t.Fatal("nil handler")
	}
}

func TestStartHTTPServer_shutdownOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := &server.Config{}
	cfg.Address = "127.0.0.1:0"

	h := initHTTPHandler(nil, initRepos(nil), zap.NewNop())

	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	err := startHTTPServer(ctx, cfg, h)
	if err != nil {
		t.Fatal(err)
	}
}
