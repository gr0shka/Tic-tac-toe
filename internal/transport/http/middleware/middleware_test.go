package middleware_test

import (
	"context"
	"encoding/base64"
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

func (m mockAuthenticateService) Register(ctx context.Context, req auth.SignUpRequest) error {
	return m.err
}

func (m mockAuthenticateService) Authenticate(ctx context.Context, login, password string) (*uuid.UUID, error) {
	return m.uID, m.err
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
			authHeader:    "Basic " + base64.StdEncoding.EncodeToString([]byte("login:password")),
			handlerCalled: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "invalid auth header",
			mockErr:       nil,
			authHeader:    "Basic " + base64.StdEncoding.EncodeToString([]byte("login")),
			handlerCalled: false,
			expectedCode:  http.StatusUnauthorized,
		},
		{
			name:          "user not exist",
			mockErr:       user.ErrUserNotFound,
			authHeader:    "Basic " + base64.StdEncoding.EncodeToString([]byte("login:password")),
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

				userStrID, ok := receivedCtx.Value(middleware.UserIDContextName).(string)
				if !ok {
					t.Errorf("expected user id to be a string got %v", receivedCtx.Value(middleware.UserIDContextName))
				}

				userID, err := uuid.Parse(userStrID)
				if err != nil {
					t.Errorf("expected user id to be a valid uuid got %v", userStrID)
				}

				if userID != newUser.ID() {
					t.Errorf("expected user id to be %v got %v", newUser.ID(), userStrID)
				}
			}
		})
	}
}
