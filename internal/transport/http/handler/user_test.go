package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
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
	type testCase struct {
		name           string
		body           dto.SignUpRequest
		mockErr        error
		expectedCode   int
		expectedErrMsg string
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
			mockErr:      user.ErrInValidLogin,
		},
		{
			name: "User already exists",

			body: dto.SignUpRequest{
				Login:    "test",
				Password: "test",
			},

			expectedCode: http.StatusBadRequest,
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

			mockUS := mockUserService{err: tc.mockErr}

			h := handler.NewUserHandler(mockUS)

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
		authHeader   string
		mockErr      error
		expectedCode int
	}

	u := user.New(uuid.New(), "login", "password")

	testCases := []testCase{
		{
			name:         "success authenticate",
			mockUser:     u,
			authHeader:   "Basic " + base64.StdEncoding.EncodeToString([]byte(u.Login()+":"+u.Password())),
			expectedCode: http.StatusOK,
			mockErr:      nil,
		},
		{
			name:         "empty authHeader",
			mockUser:     u,
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
			mockErr:      nil,
		},
		{
			name:         "invalid authHeader",
			mockUser:     u,
			authHeader:   "Basic " + base64.StdEncoding.EncodeToString([]byte(u.Login())),
			expectedCode: http.StatusUnauthorized,
			mockErr:      nil,
		},
		{
			name:         "password incorrect",
			mockUser:     u,
			authHeader:   base64.StdEncoding.EncodeToString([]byte(u.Password())),
			expectedCode: http.StatusUnauthorized,
			mockErr:      user.ErrPasswordNotMatch,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", "/user/login", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tc.authHeader)

			mockUS := mockUserService{u: tc.mockUser, err: tc.mockErr}

			h := handler.NewUserHandler(mockUS)

			h.Authenticate(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", res.StatusCode, tc.expectedCode)
			}
		})
	}

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
			expectedCode: http.StatusBadRequest,
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
