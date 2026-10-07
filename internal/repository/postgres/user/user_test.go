package user

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, cleanup, err := testutil.SetupTestDB(ctx)
	if err != nil {
		log.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	testPool = pool
	os.Exit(m.Run())
}

func truncateUsers(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE users CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}
}

func TestRepository_Save(t *testing.T) {
	truncateUsers(t)
	type testCase struct {
		name string
		u    *user.User
		err  error
	}

	userLogin := "login"
	userPassword := "password"
	newUser := user.New(uuid.New(), userLogin, userPassword)

	testCases := []testCase{
		{
			name: "success save user",
			u:    newUser,
			err:  nil,
		},
	}

	repo := New(testPool)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := repo.Save(context.Background(), tc.u)
			if err != nil {
				t.Errorf("failed to save user: %v", err)
			}

			if err != tc.err {
				t.Errorf("save user error, expected %v, got %v", tc.err, err)
			}
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	truncateUsers(t)
	type testCase struct {
		name string
		u    *user.User
		err  error
	}

	userLogin := "login"
	userPassword := "password"
	newUser := user.New(uuid.New(), userLogin, userPassword)

	testCases := []testCase{
		{
			name: "success save user and get user by id",
			u:    newUser,
			err:  nil,
		},
	}

	repo := New(testPool)
	err := repo.Save(context.Background(), newUser)
	if err != nil {
		t.Errorf("failed to save user: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			u, err := repo.GetUserByID(context.Background(), tc.u.ID())
			if err != nil {
				t.Errorf("failed to get user by id: %v", err)
			}

			if u == nil {
				t.Errorf("failed to get user by id")
			}

			if !errors.Is(err, tc.err) {
				t.Errorf("save user error, expected %v, got %v", tc.err, err)
			}

			if err == nil && u != nil {

				if u.ID() != tc.u.ID() {
					t.Errorf("save user error, expected %v, got %v", tc.u.ID(), u.ID())
				}

				if u.Login() != tc.u.Login() {
					t.Errorf("save user error, expected %v, got %v", tc.u.Login(), u.Login())
				}
			}
		})
	}
}

func TestRepository_GetUserByLogin(t *testing.T) {
	truncateUsers(t)
	type testCase struct {
		name string
		u    *user.User
		err  error
	}

	userLogin := "login"
	userPassword := "password"
	newUser := user.New(uuid.New(), userLogin, userPassword)

	testCases := []testCase{
		{
			name: "success save user and get user by login",
			u:    newUser,
			err:  nil,
		},
	}

	repo := New(testPool)
	err := repo.Save(context.Background(), newUser)
	if err != nil {
		t.Errorf("failed to save user: %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			u, err := repo.GetByLogin(context.Background(), tc.u.Login())
			if err != nil {
				t.Errorf("failed to get user by login: %v", err)
			}

			if u == nil {
				t.Errorf("failed to get user by login")
			}

			if !errors.Is(err, tc.err) {
				t.Errorf("save user error, expected %v, got %v", tc.err, err)
			}

			if err == nil && u != nil {

				if u.ID() != tc.u.ID() {
					t.Errorf("save user error, expected %v, got %v", tc.u.ID(), u.ID())
				}

				if u.Login() != tc.u.Login() {
					t.Errorf("save user error, expected %v, got %v", tc.u.Login(), u.Login())
				}
			}
		})
	}
}
