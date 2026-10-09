package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
	"golang.org/x/crypto/bcrypt"
)

type mockUserService struct {
	u   *user.User
	err error
}

func (m mockUserService) Create(ctx context.Context, login, passHash string) (*user.User, error) {
	return m.u, m.err
}

func (m mockUserService) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	return m.u, m.err
}

func (m mockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return m.u, m.err
}

func TestAuthenticateService_Authenticate(t *testing.T) {
	type testCase struct {
		name     string
		login    string
		password string
		user     *user.User
		err      error
		mockErr  error
	}

	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testPassword"), bcrypt.DefaultCost)

	testUser := user.New(
		uuid.New(),
		"testLogin",
		string(hashPass))

	testCases := []testCase{
		{
			name:     "Success authentication",
			login:    "testLogin",
			password: "testPassword",
			user:     testUser,
			err:      nil,
		},
		{
			name:     "Error authentication invalid login",
			login:    "",
			password: "testPassword",
			user:     testUser,
			err:      user.ErrInValidLogin,
		},
		{
			name:     "Error authentication invalid password",
			login:    "testLogin",
			password: "",
			user:     testUser,
			err:      user.ErrInValidPassword,
		},
		{
			name:     "Error authentication password not match",
			login:    "testLogin",
			password: "testPasswordErr",
			user:     testUser,
			err:      user.ErrPasswordNotMatch,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			us := mockUserService{tc.user, tc.mockErr}
			as := auth.NewAuthenticateService(us)

			gotUserID, err := as.Authenticate(context.Background(), tc.login, tc.password)

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
			}

			if gotUserID != nil && *gotUserID != tc.user.ID() {
				t.Errorf("gotUser.ID() = %v, want = %v", *gotUserID, tc.user.ID())
			}
		})
	}
}

func TestAuthenticateService_Register(t *testing.T) {
	type testCase struct {
		name     string
		login    string
		password string
		user     *user.User
		err      error
		mockErr  error
	}

	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testPassword"), bcrypt.DefaultCost)

	testUser := user.New(
		uuid.New(),
		"testLogin",
		string(hashPass))

	testCases := []testCase{
		{
			name:     "Error registration invalid login",
			login:    "",
			password: "testPassword",
			user:     testUser,
			err:      user.ErrInValidLogin,
			mockErr:  nil,
		},
		{
			name:     "Error registration invalid password",
			login:    "testLogin1",
			password: "",
			user:     testUser,
			err:      user.ErrInValidPassword,
			mockErr:  nil,
		},
		{
			name:     "Error registration user already exists",
			login:    "testLogin",
			password: "testPassword",
			user:     testUser,
			err:      user.ErrUserAlreadyExists,
			mockErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			us := mockUserService{tc.user, tc.mockErr}
			as := auth.NewAuthenticateService(us)

			req := auth.SignUpRequest{tc.login, tc.password}
			err := as.Register(context.Background(), req)

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
			}
		})
	}
}
