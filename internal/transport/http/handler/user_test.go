package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
)

type mockUserService struct {
	u   *user.User
	err error
}

func (m mockUserService) Register(ctx context.Context, login, password string) error {
	return m.err
}

func (m mockUserService) Authenticate(ctx context.Context, login, password string) (*user.User, error) {
	return m.u, m.err
}

func (m mockUserService) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	return m.u, m.err
}

func (m mockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return m.u, m.err
}

func TestHandler_Register(t *testing.T) {
	w := httptest.NewRecorder()

	request := dto.SignUpRequest{
		Login:    "login",
		Password: "password",
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		t.Error(err)
	}

	req, _ := http.NewRequest("POST", "/user/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	u := user.New(uuid.New(), "login", "password")
	mockUS := mockUserService{u: u, err: nil}

	h := handler.NewUserHandler(mockUS)

	h.Register(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, http.StatusOK)
	}
}

func TestHandler_Authenticate(t *testing.T) {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/user/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("login", "password")

	u := user.New(uuid.New(), "login", "password")

	mockUS := mockUserService{u: u, err: nil}

	h := handler.NewUserHandler(mockUS)

	h.Authenticate(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, http.StatusOK)
	}
}

func TestHandler_GetUserByID(t *testing.T) {
	w := httptest.NewRecorder()

	userID := uuid.New()
	req, _ := http.NewRequest("GET", "/user/"+userID.String(), nil)
	req.SetPathValue("uuid", userID.String())

	u := user.New(userID, "login", "password")
	mockUS := mockUserService{u: u, err: nil}

	h := handler.NewUserHandler(mockUS)

	h.GetUserByID(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, http.StatusOK)
	}
}
