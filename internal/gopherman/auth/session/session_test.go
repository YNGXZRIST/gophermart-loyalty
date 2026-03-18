package session

import "testing"

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		wantErr bool
	}{
		{
			name:    "valid",
			str:     "valid",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenerateToken()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
