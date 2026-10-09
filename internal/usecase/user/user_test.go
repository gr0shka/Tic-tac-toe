package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	userService "github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	user *user.User
	err  error
}

func (m MockUserRepository) Create(ctx context.Context, login, passHash string) (*user.User, error) {
	return m.user, m.err
}

func (m MockUserRepository) Save(ctx context.Context, user *user.User) error {
	return m.err
}

func (m MockUserRepository) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	return m.user, m.err
}

func (m MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return m.user, m.err
}

func TestUserService_GetByLogin(t *testing.T) {
	type testCase struct {
		name     string
		login    string
		password string
		user     *user.User
		err      error
	}

	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testPassword"), bcrypt.DefaultCost)

	testUser := user.New(
		uuid.New(),
		"testLogin",
		string(hashPass))

	testCases := []testCase{
		{
			name:     "Success get user by login",
			login:    "testLogin",
			password: "testPassword",
			user:     testUser,
			err:      nil,
		},
		{
			name:     "Error get user by login invalid login",
			login:    "",
			password: "testPassword",
			user:     testUser,
			err:      user.ErrInValidLogin,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := MockUserRepository{tc.user, nil}
			us := userService.NewUserService(repo)

			gotUser, err := us.GetByLogin(context.Background(), tc.login)

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
			}

			if gotUser != nil && gotUser.ID() != tc.user.ID() {
				t.Errorf("gotUser.ID() = %v, want = %v", gotUser.ID(), tc.user.ID())
			}

			if gotUser != nil && gotUser.Login() != tc.user.Login() {
				t.Errorf("gotUser.Login() = %v, want = %v", gotUser.Login(), tc.user.Login())
			}

			if gotUser != nil && gotUser.Password() != tc.user.Password() {
				t.Errorf("gotUser.Password() = %v, want = %v", gotUser.Password(), tc.user.Password())
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	type testCase struct {
		name     string
		login    string
		password string
		user     *user.User
		err      error
	}

	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testPassword"), bcrypt.DefaultCost)

	testUser := user.New(
		uuid.New(),
		"testLogin",
		string(hashPass))

	testCases := []testCase{
		{
			name:     "Success get user by ID",
			login:    "testLogin",
			password: "testPassword",
			user:     testUser,
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := MockUserRepository{tc.user, nil}
			us := userService.NewUserService(repo)

			gotUser, err := us.GetUserByID(context.Background(), tc.user.ID())

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
			}

			if gotUser != nil && gotUser.ID() != tc.user.ID() {
				t.Errorf("gotUser.ID() = %v, want = %v", gotUser.ID(), tc.user.ID())
			}

			if gotUser != nil && gotUser.Login() != tc.user.Login() {
				t.Errorf("gotUser.Login() = %v, want = %v", gotUser.Login(), tc.user.Login())
			}

			if gotUser != nil && gotUser.Password() != tc.user.Password() {
				t.Errorf("gotUser.Password() = %v, want = %v", gotUser.Password(), tc.user.Password())
			}
		})
	}
}
