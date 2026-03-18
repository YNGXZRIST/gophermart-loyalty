package contextkey

import (
	"context"
	"reflect"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		args       args
		wantUserID int64
		wantOk     bool
	}{
		{
			name: "success",
			args: args{
				ctx: context.WithValue(context.Background(), userIDKey, int64(1)),
			},
			wantUserID: 1,
			wantOk:     true,
		},
		{
			name: "failure",
			args: args{
				ctx: context.Background(),
			},
			wantUserID: 0,
			wantOk:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, gotOk := UserIDFromContext(tt.args.ctx)
			if gotUserID != tt.wantUserID {
				t.Errorf("UserIDFromContext() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
			if gotOk != tt.wantOk {
				t.Errorf("UserIDFromContext() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestWithUserID(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID int64
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			want: context.WithValue(context.Background(), userIDKey, int64(1)),
		},
		{
			name: "zero uid",
			args: args{
				ctx:    context.Background(),
				userID: 0,
			},
			want: context.WithValue(context.Background(), userIDKey, int64(0)),
		},
		{
			name: "negative uid",
			args: args{
				ctx:    context.Background(),
				userID: -1,
			},
			want: context.WithValue(context.Background(), userIDKey, int64(-1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithUserID(tt.args.ctx, tt.args.userID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}
