package validator

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeAndValidate(t *testing.T) {
	type args struct {
		r *http.Request
		v Validator
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				r: func() *http.Request {
					body := strings.NewReader(`{"login":"u","password":"p"}`)
					r := httptest.NewRequest(http.MethodPost, "/", body)
					return r
				}(),
				v: &testUser{},
			},
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				r: func() *http.Request {
					body := strings.NewReader(`{"loword":"p"}`)
					r := httptest.NewRequest(http.MethodPost, "/", body)
					return r
				}(),
				v: &testUser{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DecodeAndValidate(tt.args.r, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("DecodeAndValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (t *testUser) Validate() error {
	if t.Login == "" {
		return errors.New("invalid login")
	}
	if t.Password == "" {
		return errors.New("invalid password")
	}
	return nil
}
