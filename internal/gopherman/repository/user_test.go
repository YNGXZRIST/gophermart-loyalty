package repository

import (
	"context"
	"testing"

	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository/mock"

	"github.com/golang/mock/gomock"
)

func TestUserRepository_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)

	repos := Repositories{
		User: mockUser,
	}
	_ = repos
}

func TestUserRepository_GetByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)
	repos := Repositories{User: mockUser}
	ctx := context.Background()

	u := &model.User{ID: 1, Login: "alice"}
	mockUser.EXPECT().
		GetByLogin(gomock.Any(), "alice").
		Return(u, nil)

	got, err := repos.User.GetByLogin(ctx, "alice")
	if err != nil {
		t.Fatalf("GetByLogin: %v", err)
	}
	if got != u || got.Login != "alice" {
		t.Errorf("GetByLogin: got %v, want user alice", got)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)
	repos := Repositories{User: mockUser}
	ctx := context.Background()
	userID := int64(42)

	u := &model.User{ID: userID, Login: "bob"}
	mockUser.EXPECT().
		GetByID(gomock.Any(), userID).
		Return(u, nil)

	got, err := repos.User.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != userID || got.Login != "bob" {
		t.Errorf("GetByID: got %v, want user id=42 bob", got)
	}
}

func TestUserRepository_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := mock.NewMockUserRepository(ctrl)
	repos := Repositories{User: mockUser}
	ctx := context.Background()

	u := &model.User{ID: 3, Login: "newuser"}
	mockUser.EXPECT().
		Register(gomock.Any(), "newuser", "secret", "1.2.3.4").
		Return(u, nil)

	got, err := repos.User.Register(ctx, "newuser", "secret", "1.2.3.4")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if got.ID != 3 || got.Login != "newuser" {
		t.Errorf("Register: got %v, want id=3 newuser", got)
	}
}
