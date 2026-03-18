package model

import "testing"

func TestRegisterRequest_Validate(t *testing.T) {
	type fields struct {
		Login string
		Pass  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Login: "12345",
				Pass:  "123456",
			},
			wantErr: false,
		},
		{
			name: "invalid login",
			fields: fields{
				Login: "1",
				Pass:  "123456",
			},
			wantErr: true,
		},
		{
			name: "invalid pass",
			fields: fields{
				Login: "12345",
				Pass:  "123",
			},
			wantErr: true,
		},
		{
			name: "empty login",
			fields: fields{
				Login: "",
				Pass:  "123456",
			},
			wantErr: true,
		},
		{
			name: "empty pass",
			fields: fields{
				Login: "12345",
				Pass:  "",
			},
			wantErr: true,
		},
		{
			name: "empty all",
			fields: fields{
				Login: "",
				Pass:  "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RegisterRequest{
				Login: tt.fields.Login,
				Pass:  tt.fields.Pass,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
