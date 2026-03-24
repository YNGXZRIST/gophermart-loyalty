package logger

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		mode    string
		cmdType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "initialize logger in dev mode for server",
			args: args{
				mode:    TypeModeTest,
				cmdType: ServerLgr,
			},
			wantErr: false,
		},
		{
			name: "initialize logger in prod mode for server",
			args: args{
				mode:    TypeModeProduction,
				cmdType: ServerLgr,
			},
			wantErr: false,
		},
		{
			name: "initialize logger with invalid mode",
			args: args{
				mode:    "invalid_mode",
				cmdType: ServerLgr,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Initialize(tt.args.mode, tt.args.cmdType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Initialize() got = nil, want non-nil logger")
			}
		})
	}
}

func Test_createProductionLogger(t *testing.T) {
	type args struct {
		cmdType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create production logger for server",
			args: args{
				cmdType: ServerLgr,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createProductionLogger(tt.args.cmdType)
			if (err != nil) != tt.wantErr {
				t.Errorf("createProductionLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("createProductionLogger() got = nil, want non-nil logger")
			}
		})
	}
}

func Test_createDevelopmentLogger(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create development logger",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createDevelopmentLogger()
			if (err != nil) != tt.wantErr {
				t.Errorf("createDevelopmentLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("createDevelopmentLogger() got = nil, want non-nil logger")
			}
		})
	}
}
