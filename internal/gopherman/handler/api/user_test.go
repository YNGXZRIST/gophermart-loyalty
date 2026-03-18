package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"gophermart-loyalty/internal/gopherman/auth/password"
	"gophermart-loyalty/internal/gopherman/model"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := int64(10)
	login := "test"
	ip := "127.0.0.1"
	token := "tokenABC"

	hash, err := password.Hash("correct-pass")
	if err != nil {
		t.Fatalf("password.Hash() error: %v", err)
	}

	mockUser := mock.NewMockUserRepository(ctrl)
	mockUser.EXPECT().
		GetByLogin(gomock.Any(), login).
		Return(&model.User{ID: userID, Login: login, Pass: hash}, nil)
	mockUser.EXPECT().
		CreateSession(gomock.Any(), userID, ip).
		Return(token, nil)

	D, _ := newMockConnDB(t)
	handler := NewHandler(D, repository.Repositories{User: mockUser}, zap.NewNop())

	reqBody, _ := json.Marshal(model.RegisterRequest{Login: login, Pass: "correct-pass"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.RemoteAddr = ip + ":1234"

	w := httptest.NewRecorder()
	handler.Login(w, req.WithContext(context.Background()))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("Login status code = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Authorization"), "Bearer "+token; got != want {
		t.Fatalf("Authorization header = %q, want %q", got, want)
	}
}

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := int64(7)
	login := "test2"
	ip := "127.0.0.1"
	token := "tokenABC"

	mockUser := mock.NewMockUserRepository(ctrl)
	mockUser.EXPECT().
		GetByLogin(gomock.Any(), login).
		Return(nil, sql.ErrNoRows)
	mockUser.EXPECT().
		Register(gomock.Any(), login, "secret-pass", ip).
		Return(&model.User{ID: userID, Login: login}, nil)
	mockUser.EXPECT().
		CreateSession(gomock.Any(), userID, ip).
		Return(token, nil)

	D, _ := newMockConnDB(t)
	handler := NewHandler(D, repository.Repositories{User: mockUser}, zap.NewNop())

	reqBody, _ := json.Marshal(model.RegisterRequest{Login: login, Pass: "secret-pass"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.RemoteAddr = ip + ":1234"

	w := httptest.NewRecorder()
	handler.Register(w, req.WithContext(context.Background()))

	if got, want := w.Code, http.StatusOK; got != want {
		t.Fatalf("Register status code = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Authorization"), "Bearer "+token; got != want {
		t.Fatalf("Authorization header = %q, want %q", got, want)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not-json")))
	req2.RemoteAddr = ip + ":1234"
	handler.Register(w2, req2.WithContext(context.Background()))
	if got, want := w2.Code, http.StatusBadRequest; got != want {
		t.Fatalf("Register(bad json) status code = %d, want %d", got, want)
	}
}
