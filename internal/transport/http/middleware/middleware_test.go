package middleware_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
)

type mockAuthenticateService struct {
	uID *uuid.UUID
	err error
}

func (m mockAuthenticateService) Authenticate(ctx context.Context, login, password string) (*auth.JwtResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockAuthenticateService) AuthenticateByToken(token string) (uuid.UUID, error) {
	return *m.uID, m.err
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

func TestAuthenticator(t *testing.T) {
	type testCase struct {
		name          string
		mockErr       error
		mockUUID      uuid.UUID
		authHeader    string
		handlerCalled bool
		expectedCode  int
	}

	testCases := []testCase{
		{
			name:          "success authenticate",
			mockErr:       nil,
			authHeader:    "Bearer " + base64.StdEncoding.EncodeToString([]byte("login:password")),
			handlerCalled: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "invalid auth header",
			mockErr:       errors.New("invalid auth header"),
			authHeader:    "Bearer " + base64.StdEncoding.EncodeToString([]byte("login")),
			handlerCalled: false,
			expectedCode:  http.StatusUnauthorized,
		},
		{
			name:          "user not exist",
			mockErr:       user.ErrUserNotFound,
			authHeader:    "Bearer " + base64.StdEncoding.EncodeToString([]byte("login:password")),
			handlerCalled: false,
			expectedCode:  http.StatusUnauthorized,
		},
	}

	newUser := user.New(uuid.New(), "login", "password")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockAS := &mockAuthenticateService{new(newUser.ID()), tc.mockErr}
			ua := middleware.NewUserAuthenticator(mockAS)

			var (
				gotHandlerCalled bool
				receivedCtx      context.Context
			)
			handle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedCtx = r.Context()
				gotHandlerCalled = true
			})

			handleWithMiddleware := ua.Authenticate(handle)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Authorization", tc.authHeader)

			handleWithMiddleware.ServeHTTP(w, r)

			if tc.handlerCalled != gotHandlerCalled {
				t.Errorf("expected handler to be called %v got %v", tc.handlerCalled, tc.handlerCalled)
			}

			if tc.expectedCode != w.Code {
				t.Errorf("expected status code to be %d got %d", tc.expectedCode, w.Code)
			}

			if gotHandlerCalled {

				userID, ok := middleware.UserIDFromContext(receivedCtx)
				if !ok {
					t.Errorf("expected user id to be a valid uuid got %v", userID)
				}

				if userID != newUser.ID() {
					t.Errorf("expected user id to be %v got %v", newUser.ID(), userID)
				}
			}
		})
	}
}

func TestCORS(t *testing.T) {
	type testCase struct {
		name                string
		method              string
		origin              string
		expectHandlerCalled bool
		expectCode          int
		expectAllowOrigin   string
		expectAllowCreds    string
		expectAllowMethods  string
		expectAllowHeaders  string
	}

	testCases := []testCase{
		{
			name:                "allowed origin GET request",
			method:              http.MethodGet,
			origin:              "http://localhost:3000",
			expectHandlerCalled: true,
			expectCode:          http.StatusOK,
			expectAllowOrigin:   "http://localhost:3000",
			expectAllowCreds:    "true",
		},
		{
			name:                "disallowed origin request",
			method:              http.MethodGet,
			origin:              "http://untrusted-site.com",
			expectHandlerCalled: true,
			expectCode:          http.StatusOK,
			expectAllowOrigin:   "",
			expectAllowCreds:    "",
		},
		{
			name:                "OPTIONS preflight request",
			method:              http.MethodOptions,
			origin:              "http://localhost:3000",
			expectHandlerCalled: false,
			expectCode:          http.StatusNoContent,
			expectAllowOrigin:   "http://localhost:3000",
			expectAllowCreds:    "true",
			expectAllowMethods:  "GET, POST, OPTIONS",
			expectAllowHeaders:  "Origin, X-Requested-With, Content-Type, Accept, Authorization",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			corsHandler := middleware.CORS(next)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/test", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			corsHandler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			if tc.expectHandlerCalled != handlerCalled {
				t.Errorf("expected handlerCalled %v, got %v", tc.expectHandlerCalled, handlerCalled)
			}

			if res.StatusCode != tc.expectCode {
				t.Errorf("expected status %d, got %d", tc.expectCode, res.StatusCode)
			}

			if res.Header.Get("Vary") != "Origin" {
				t.Errorf("expected Vary 'Origin', got '%s'", res.Header.Get("Vary"))
			}

			if res.Header.Get("Access-Control-Allow-Origin") != tc.expectAllowOrigin {
				t.Errorf("expected Access-Control-Allow-Origin '%s', got '%s'",
					tc.expectAllowOrigin, res.Header.Get("Access-Control-Allow-Origin"))
			}

			if res.Header.Get("Access-Control-Allow-Credentials") != tc.expectAllowCreds {
				t.Errorf("expected Access-Control-Allow-Credentials '%s', got '%s'",
					tc.expectAllowCreds, res.Header.Get("Access-Control-Allow-Credentials"))
			}

			if tc.expectAllowMethods != "" && res.Header.Get("Access-Control-Allow-Methods") != tc.expectAllowMethods {
				t.Errorf("expected Access-Control-Allow-Methods '%s', got '%s'",
					tc.expectAllowMethods, res.Header.Get("Access-Control-Allow-Methods"))
			}

			if tc.expectAllowHeaders != "" && res.Header.Get("Access-Control-Allow-Headers") != tc.expectAllowHeaders {
				t.Errorf("expected Access-Control-Allow-Headers '%s', got '%s'",
					tc.expectAllowHeaders, res.Header.Get("Access-Control-Allow-Headers"))
			}
		})
	}
}

func TestRecovery(t *testing.T) {
	t.Run("recovers from panic and returns 500 JSON", func(t *testing.T) {
		panickingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("unexpected runtime error")
		})

		recoveryHandler := middleware.Recovery(panickingHandler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/panic-route", nil)

		recoveryHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
		}

		if res.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", res.Header.Get("Content-Type"))
		}

		body := w.Body.String()
		expectedBody := `{"error":"internal server error"}`
		if body != expectedBody {
			t.Errorf("expected body '%s', got '%s'", expectedBody, body)
		}
	})

	t.Run("passes through without panic", func(t *testing.T) {
		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})

		recoveryHandler := middleware.Recovery(normalHandler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/normal-route", nil)

		recoveryHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}

		if w.Body.String() != "ok" {
			t.Errorf("expected body 'ok', got '%s'", w.Body.String())
		}
	})
}

func TestRequestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	})

	loggedHandler := middleware.RequestLogger(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/logged-endpoint", nil)

	loggedHandler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, res.StatusCode)
	}

	if w.Body.String() != "accepted" {
		t.Errorf("expected body 'accepted', got '%s'", w.Body.String())
	}
}
