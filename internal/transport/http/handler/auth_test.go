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
	u       *uuid.UUID
	err     error
	jwtResp *auth.JwtResponse
}

func (m mockAuthenticateService) Authenticate(ctx context.Context, login, password string) (*auth.JwtResponse, error) {
	if m.jwtResp != nil {
		return m.jwtResp, m.err
	}
	return &auth.JwtResponse{
		Type:                  "correct",
		AccessToken:           "correct",
		RefreshToken:          "correct",
		RefreshTokenExpiresAt: time.Now(),
	}, m.err
}

func (m mockAuthenticateService) AuthenticateByToken(token string) (uuid.UUID, error) {
	if m.u != nil {
		return *m.u, m.err
	}
	return uuid.Nil, m.err
}

func (m mockAuthenticateService) RefreshAccessToken(ctx context.Context, refreshToken string) (*auth.JwtResponse, error) {
	return m.jwtResp, m.err
}

func (m mockAuthenticateService) RefreshRefreshToken(ctx context.Context, refreshToken string) (*auth.JwtResponse, error) {
	return m.jwtResp, m.err
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

			mockAS := mockAuthenticateService{err: tc.mockErr}

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

			mockAS := mockAuthenticateService{err: tc.mockErr}

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

func TestHandler_RefreshAccessToken(t *testing.T) {
	jwtSuccess := &auth.JwtResponse{
		Type:         "Bearer",
		AccessToken:  "new_access_token_123",
		RefreshToken: "refresh_token_123",
	}

	type testCase struct {
		name         string
		cookieToken  string
		bodyToken    string
		mockResp     *auth.JwtResponse
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success refresh via cookie",
			cookieToken:  "cookie_refresh_token",
			mockResp:     jwtSuccess,
			expectedCode: http.StatusOK,
		},
		{
			name:         "success refresh via json body",
			bodyToken:    "body_refresh_token",
			mockResp:     jwtSuccess,
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed missing token in cookie and body",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "failed invalid refresh token in service",
			cookieToken:  "invalid_refresh_token",
			mockErr:      user.ErrInvalidCredentials,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "failed internal service error",
			cookieToken:  "cookie_refresh_token",
			mockErr:      errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			var body bytes.Buffer
			if tc.bodyToken != "" {
				_ = json.NewEncoder(&body).Encode(dto.RefreshJwtRequest{RefreshToken: tc.bodyToken})
			}

			req := httptest.NewRequest(http.MethodPost, "/user/token", &body)
			if tc.cookieToken != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tc.cookieToken})
			}

			mockAS := mockAuthenticateService{jwtResp: tc.mockResp, err: tc.mockErr}
			h := handler.NewAuthHandler(mockAS)

			h.RefreshAccessToken(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				var response dto.JwtResponse
				if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
					t.Fatalf("failed decoding response: %v", err)
				}
				if response.AccessToken != tc.mockResp.AccessToken {
					t.Errorf("expected AccessToken %s, got %s", tc.mockResp.AccessToken, response.AccessToken)
				}
				if response.RefreshToken != tc.mockResp.RefreshToken {
					t.Errorf("expected RefreshToken %s, got %s", tc.mockResp.RefreshToken, response.RefreshToken)
				}
				if response.Type != tc.mockResp.Type {
					t.Errorf("expected Type %s, got %s", tc.mockResp.Type, response.Type)
				}
			}
		})
	}
}

func TestHandler_RefreshRefreshToken(t *testing.T) {
	expireTime := time.Now().Add(time.Hour * 24 * 7)
	jwtSuccess := &auth.JwtResponse{
		Type:                  "Bearer",
		AccessToken:           "new_access_token_456",
		RefreshToken:          "rotated_refresh_token_789",
		RefreshTokenExpiresAt: expireTime,
	}

	type testCase struct {
		name         string
		cookieToken  string
		bodyToken    string
		mockResp     *auth.JwtResponse
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success refresh refresh token (sets cookie)",
			cookieToken:  "old_cookie_refresh_token",
			mockResp:     jwtSuccess,
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed missing token",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "failed service error",
			cookieToken:  "old_token",
			mockErr:      user.ErrInvalidCredentials,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			var body bytes.Buffer
			if tc.bodyToken != "" {
				_ = json.NewEncoder(&body).Encode(dto.RefreshJwtRequest{RefreshToken: tc.bodyToken})
			}

			req := httptest.NewRequest(http.MethodPost, "/user/refresh", &body)
			if tc.cookieToken != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tc.cookieToken})
			}

			mockAS := mockAuthenticateService{jwtResp: tc.mockResp, err: tc.mockErr}
			h := handler.NewAuthHandler(mockAS)

			h.RefreshRefreshToken(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				var response dto.JwtResponse
				if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
					t.Fatalf("failed decoding response: %v", err)
				}
				if response.AccessToken != tc.mockResp.AccessToken {
					t.Errorf("expected AccessToken %s, got %s", tc.mockResp.AccessToken, response.AccessToken)
				}
				if response.RefreshToken != tc.mockResp.RefreshToken {
					t.Errorf("expected RefreshToken %s, got %s", tc.mockResp.RefreshToken, response.RefreshToken)
				}

				cookies := res.Cookies()
				var foundCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "refresh_token" {
						foundCookie = c
						break
					}
				}
				if foundCookie == nil {
					t.Fatal("expected Set-Cookie header with refresh_token, got none")
				}
				if foundCookie.Value != tc.mockResp.RefreshToken {
					t.Errorf("expected cookie value %s, got %s", tc.mockResp.RefreshToken, foundCookie.Value)
				}
			}
		})
	}
}
