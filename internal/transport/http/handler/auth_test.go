package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
)

type mockAuthenticateService struct {
	u   *uuid.UUID
	err error
}

func (m mockAuthenticateService) Authenticate(ctx context.Context, login, password string) (*auth.JwtResponse, error) {
	return &auth.JwtResponse{
		Type:                  "correct",
		AccessToken:           "correct",
		RefreshToken:          "correct",
		RefreshTokenExpiresAt: time.Now(),
	}, m.err
}

func (m mockAuthenticateService) AuthenticateByToken(token string) (uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockAuthenticateService) RefreshAccessToken(ctx context.Context, refreshToken string) (*auth.JwtResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockAuthenticateService) RefreshRefreshToken(ctx context.Context, refreshToken string) (*auth.JwtResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockAuthenticateService) Register(ctx context.Context, req auth.SignUpRequest) error {
	return m.err
}

func TestHandler_Register(t *testing.T) {
	type testCase struct {
		name         string
		body         dto.SignUpRequest
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name: "success registration",

			body: dto.SignUpRequest{
				Login:    "test",
				Password: "test",
			},

			expectedCode: http.StatusCreated,
			mockErr:      nil,
		},
		{
			name: "Empty login",

			body: dto.SignUpRequest{
				Login:    "",
				Password: "test",
			},

			expectedCode: http.StatusBadRequest,
			mockErr:      user.ErrInvalidArgument,
		},
		{
			name: "User already exists",

			body: dto.SignUpRequest{
				Login:    "test",
				Password: "test",
			},

			expectedCode: http.StatusConflict,
			mockErr:      user.ErrUserAlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			requestBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Error(err)
			}

			req, _ := http.NewRequest("POST", "/user/register", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			mockAS := mockAuthenticateService{nil, tc.mockErr}

			h := handler.NewAuthHandler(mockAS)

			h.Register(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, tc.expectedCode)
			}
		})
	}
}

func TestHandler_Authenticate(t *testing.T) {
	type testCase struct {
		name         string
		mockUser     *user.User
		authBody     string
		mockErr      error
		expectedCode int
	}

	u := user.New(uuid.New(), "login", "password")

	correctBody, _ := json.Marshal(dto.JwtRequest{
		Login:    "login",
		Password: "password",
	})

	unCorrectPasswordBody, _ := json.Marshal(dto.JwtRequest{
		Login:    "login",
		Password: "uncorrected",
	})

	testCases := []testCase{
		{
			name:         "success authenticate",
			mockUser:     u,
			authBody:     string(correctBody),
			expectedCode: http.StatusOK,
			mockErr:      nil,
		},
		{
			name:         "empty authBody",
			mockUser:     u,
			authBody:     "",
			expectedCode: http.StatusBadRequest,
			mockErr:      errors.New("empty authBody"),
		},
		{
			name:         "password incorrect",
			mockUser:     u,
			authBody:     string(unCorrectPasswordBody),
			expectedCode: http.StatusUnauthorized,
			mockErr:      user.ErrInvalidCredentials,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(tc.authBody))
			req.Header.Set("Content-Type", "application/json")

			mockAS := mockAuthenticateService{nil, tc.mockErr}

			h := handler.NewAuthHandler(mockAS)

			h.Authenticate(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, tc.expectedCode)
			}
		})
	}

}
