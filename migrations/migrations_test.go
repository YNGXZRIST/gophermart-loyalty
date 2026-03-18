package migrations

import (
	"strings"
	"testing"
)

func TestMigrate(t *testing.T) {
	tests := []struct {
		name            string
		dsn             string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "empty DSN returns error",
			dsn:             "",
			wantErr:         true,
			wantErrContains: "database DSN is not set",
		},
		{
			name:            "DSN contains invalid characters",
			dsn:             "123",
			wantErr:         true,
			wantErrContains: "open db:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Migrate(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.wantErrContains != "" {
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Migrate() error = %q, want containing %q", err.Error(), tt.wantErrContains)
				}
			}
		})
	}
}
