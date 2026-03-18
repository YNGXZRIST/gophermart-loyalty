//go:build integration
// +build integration

package migrations

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestMigrate_integration(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start postgres: %v", err)
	}
	defer func() { _ = pool.Purge(resource) }()
	_ = resource.Expire(30)

	hostPort := resource.GetPort("5432/tcp")
	dsn := "postgres://postgres:postgres@localhost:" + hostPort + "/postgres?sslmode=disable"

	if err = pool.Retry(func() error {
		conn, retryErr := sql.Open("pgx", dsn)
		if retryErr != nil {
			return retryErr
		}
		defer conn.Close()
		return conn.Ping()
	}); err != nil {
		t.Fatalf("Postgres not ready: %v", err)
	}

	if err := Migrate(dsn); err != nil {
		t.Fatalf("Migrate() = %v", err)
	}

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Open after migrate: %v", err)
	}
	defer conn.Close()

	var exists bool
	err = conn.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'metrics'
		)`).Scan(&exists)
	if err != nil {
		t.Fatalf("Check metrics table: %v", err)
	}
	if !exists {
		t.Error("After Migrate(): table metrics does not exist")
	}
}
