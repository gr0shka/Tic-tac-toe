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
			err:      user.ErrInvalidArgument,
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

func TestUserService_Create(t *testing.T) {
	dbErr := errors.New("db error")

	type testCase struct {
		name     string
		login    string
		passHash string
		repoErr  error
		wantErr  error
	}

	testCases := []testCase{
		{
			name:     "Success create user",
			login:    "newUser",
			passHash: "hashedPassword123",
			repoErr:  nil,
			wantErr:  nil,
		},
		{
			name:     "Failed repo save error",
			login:    "newUser",
			passHash: "hashedPassword123",
			repoErr:  dbErr,
			wantErr:  dbErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := MockUserRepository{err: tc.repoErr}
			us := userService.NewUserService(repo)

			gotUser, err := us.Create(context.Background(), tc.login, tc.passHash)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Create() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if gotUser == nil {
					t.Fatal("Create() returned nil user, want non-nil")
				}
				if gotUser.ID() == uuid.Nil {
					t.Errorf("Create() gotUser.ID() = uuid.Nil, want valid UUID")
				}
				if gotUser.Login() != tc.login {
					t.Errorf("Create() gotUser.Login() = %v, want %v", gotUser.Login(), tc.login)
				}
				if gotUser.Password() != tc.passHash {
					t.Errorf("Create() gotUser.Password() = %v, want %v", gotUser.Password(), tc.passHash)
				}
			}
		})
	}
}
