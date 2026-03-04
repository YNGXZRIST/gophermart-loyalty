package internal

import (
	"reflect"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")

	t.Run("success with flags", func(t *testing.T) {
		got, err := NewOptions([]string{"-a", ":8080", "-d", "postgres://localhost/gophermart"})
		if err != nil {
			t.Errorf("NewOptions() error = %v", err)
			return
		}
		if got.Address != ":8080" || got.DatabaseURL != "postgres://localhost/gophermart" {
			t.Errorf("NewOptions() got = %+v", got)
		}
	})

	t.Run("env overrides flags", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", "127.0.0.1:8080")
		t.Setenv("DATABASE_URI", "postgres://env/db")
		got, err := NewOptions([]string{"-a", "ignored", "-d", "ignored"})
		if err != nil {
			t.Errorf("NewOptions() error = %v", err)
			return
		}
		if got.Address != "127.0.0.1:8080" || got.DatabaseURL != "postgres://env/db" {
			t.Errorf("NewOptions() env override: got Address=%q DatabaseURL=%q", got.Address, got.DatabaseURL)
		}
	})

	t.Run("invalid flag", func(t *testing.T) {
		_, err := NewOptions([]string{"-invalid"})
		if err == nil {
			t.Error("NewOptions() expected error for unknown flag")
		}
	})
}

func TestOptions_ValidateOptions(t *testing.T) {
	type fields struct {
		Address        string
		DatabaseURL    string
		AccrualAddress string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid",
			fields:  fields{Address: "localhost:8080", DatabaseURL: "postgres://localhost/db", AccrualAddress: "http://accrual:8080"},
			wantErr: false,
		},
		{
			name:    "empty address",
			fields:  fields{Address: "", DatabaseURL: "postgres://localhost/db", AccrualAddress: "http://accrual:8080"},
			wantErr: true,
		},
		{
			name:    "empty database URL",
			fields:  fields{Address: "localhost:8080", DatabaseURL: "", AccrualAddress: "http://accrual:8080"},
			wantErr: true,
		},
		{
			name:    "empty accrual address",
			fields:  fields{Address: "localhost:8080", DatabaseURL: "postgres://localhost/db", AccrualAddress: ""},
			wantErr: true,
		},
		{
			name:    "all empty",
			fields:  fields{Address: "", DatabaseURL: "", AccrualAddress: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &Options{
				Address:        tt.fields.Address,
				DatabaseURL:    tt.fields.DatabaseURL,
				AccrualAddress: tt.fields.AccrualAddress,
			}
			if err := opt.ValidateOptions(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOptions_parseEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "")
	t.Setenv("DATABASE_URI", "")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "")

	t.Run("empty env does not error", func(t *testing.T) {
		opt := &Options{Address: "localhost", DatabaseURL: "postgres://local", AccrualAddress: ""}
		if err := opt.parseEnv(); err != nil {
			t.Errorf("parseEnv() unexpected error = %v", err)
		}
	})

	t.Run("env overrides values", func(t *testing.T) {
		t.Setenv("RUN_ADDRESS", "0.0.0.0:8080")
		t.Setenv("DATABASE_URI", "postgres://env/db")
		t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual:8080")
		opt := &Options{}
		if err := opt.parseEnv(); err != nil {
			t.Errorf("parseEnv() error = %v", err)
		}
		if opt.Address != "0.0.0.0:8080" || opt.DatabaseURL != "postgres://env/db" || opt.AccrualAddress != "http://accrual:8080" {
			t.Errorf("parseEnv() got Address=%q DatabaseURL=%q AccrualAddress=%q", opt.Address, opt.DatabaseURL, opt.AccrualAddress)
		}
	})
}

func Test_parseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *Options
		wantErr bool
	}{
		{
			name:    "defaults",
			args:    []string{},
			want:    &Options{Address: "localhost", DatabaseURL: "", AccrualAddress: ""},
			wantErr: false,
		},
		{
			name:    "flag -a",
			args:    []string{"-a", ":8080"},
			want:    &Options{Address: ":8080", DatabaseURL: "", AccrualAddress: ""},
			wantErr: false,
		},
		{
			name:    "flag -d",
			args:    []string{"-d", "postgres://localhost/db"},
			want:    &Options{Address: "localhost", DatabaseURL: "postgres://localhost/db", AccrualAddress: ""},
			wantErr: false,
		},
		{
			name:    "flag -db (currently binds to DatabaseURL)",
			args:    []string{"-db", "postgres://other/db"},
			want:    &Options{Address: "localhost", DatabaseURL: "postgres://other/db", AccrualAddress: ""},
			wantErr: false,
		},
		{
			name:    "all flags",
			args:    []string{"-a", "0.0.0.0:8080", "-d", "postgres://user:pass@host/db"},
			want:    &Options{Address: "0.0.0.0:8080", DatabaseURL: "postgres://user:pass@host/db", AccrualAddress: ""},
			wantErr: false,
		},
		{
			name:    "unknown flag",
			args:    []string{"-unknown", "value"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
