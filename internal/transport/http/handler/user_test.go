package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
)

type mockUserService struct {
	u   *user.User
	err error
}

func (m mockUserService) Create(ctx context.Context, login, passHash string) (*user.User, error) {
	return m.u, m.err
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

func TestHandler_GetUserByID(t *testing.T) {
	type testCase struct {
		name         string
		mockUser     *user.User
		mockErr      error
		expectedCode int
		pathUUID     string
	}

	u := user.New(uuid.New(), "login", "password")

	testCases := []testCase{
		{
			name:         "success get user by id",
			mockUser:     u,
			mockErr:      nil,
			expectedCode: http.StatusOK,
			pathUUID:     u.ID().String(),
		},
		{
			name:         "invalid uuid",
			mockUser:     u,
			mockErr:      nil,
			expectedCode: http.StatusBadRequest,
			pathUUID:     "",
		},
		{
			name:         "user not found",
			mockUser:     u,
			mockErr:      user.ErrUserNotFound,
			expectedCode: http.StatusNotFound,
			pathUUID:     u.ID().String(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", "/user/"+tc.pathUUID, nil)
			req.SetPathValue("uuid", tc.pathUUID)

			mockUS := mockUserService{u: tc.mockUser, err: tc.mockErr}

			h := handler.NewUserHandler(mockUS)

			h.GetUserByID(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, tc.expectedCode)
			}

			if res.StatusCode == http.StatusOK {
				var dtoUserResp dto.UserInfoResponse

				err := json.NewDecoder(res.Body).Decode(&dtoUserResp)
				if err != nil {
					t.Error(err)
				}

				if dtoUserResp.ID != tc.mockUser.ID() {
					t.Errorf("expected %v, got %v", tc.mockUser.ID(), dtoUserResp.ID)
				}

				if dtoUserResp.Login != tc.mockUser.Login() {
					t.Errorf("expected %v, got %v", tc.mockUser.Login(), dtoUserResp.Login)
				}
			}
		})
	}
}

func TestHandler_GetUserByToken(t *testing.T) {
	u := user.New(uuid.New(), "testUser", "secretHash")

	type testCase struct {
		name         string
		setAuthUser  bool
		mockUser     *user.User
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success get user by token",
			setAuthUser:  true,
			mockUser:     u,
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed user not in context",
			setAuthUser:  false,
			mockUser:     nil,
			mockErr:      nil,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "user not found",
			setAuthUser:  true,
			mockUser:     nil,
			mockErr:      user.ErrUserNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "failed internal service error",
			setAuthUser:  true,
			mockUser:     nil,
			mockErr:      errors.New("db connection failure"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/user/me", nil)

			if tc.setAuthUser {
				ctx := middleware.ContextWithUserID(req.Context(), u.ID())
				req = req.WithContext(ctx)
			}

			mockUS := mockUserService{u: tc.mockUser, err: tc.mockErr}
			h := handler.NewUserHandler(mockUS)

			h.GetUserByToken(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", res.StatusCode, tc.expectedCode)
			}

			if res.StatusCode == http.StatusOK {
				var dtoUserResp dto.UserInfoResponse
				if err := json.NewDecoder(res.Body).Decode(&dtoUserResp); err != nil {
					t.Fatalf("failed decoding response body: %v", err)
				}

				if dtoUserResp.ID != tc.mockUser.ID() {
					t.Errorf("expected ID %v, got %v", tc.mockUser.ID(), dtoUserResp.ID)
				}

				if dtoUserResp.Login != tc.mockUser.Login() {
					t.Errorf("expected Login %v, got %v", tc.mockUser.Login(), dtoUserResp.Login)
				}
			}
		})
	}
}
