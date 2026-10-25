package token

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	auth2 "github.com/gr0shka/Tic-tac-toe/internal/domain/auth"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/testutil"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
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

func truncateTables(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE users CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func createTestUser(t *testing.T, id uuid.UUID, login string) {
	t.Helper()
	_, err := testPool.Exec(
		context.Background(),
		"INSERT INTO users (id, login, password) VALUES ($1, $2, $3)",
		id, login, "hashed_secret",
	)
	if err != nil {
		t.Fatalf("failed inserting test user: %v", err)
	}
}

func TestRepository_SaveRefreshToken(t *testing.T) {
	truncateTables(t)

	userID := uuid.New()
	createTestUser(t, userID, "user_save")

	now := time.Now().Truncate(time.Microsecond)
	validToken := auth2.NewTokenWithFields(uuid.New(), userID, "", now.Add(time.Hour), now)
	duplicateIDToken := auth2.NewTokenWithFields(validToken.ID(), userID, "", now.Add(time.Hour*2), now)
	missingUserToken := auth2.NewTokenWithFields(uuid.New(), uuid.New(), "", now.Add(time.Hour), now)

	type testCase struct {
		name    string
		token   *auth2.Token
		wantErr bool
	}

	testCases := []testCase{
		{
			name:    "success save refresh token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "failed duplicate token id",
			token:   duplicateIDToken,
			wantErr: true,
		},
		{
			name:    "failed foreign key violation non-existent user",
			token:   missingUserToken,
			wantErr: true,
		},
	}

	repo := NewTokenRepository(testPool)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.SaveRefreshToken(context.Background(), *tc.token)
			if (err != nil) != tc.wantErr {
				t.Fatalf("SaveRefreshToken() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestRepository_GetRefreshTokenFromID(t *testing.T) {
	truncateTables(t)

	userID := uuid.New()
	createTestUser(t, userID, "user_get")

	now := time.Now().Truncate(time.Microsecond)
	existingToken := auth2.NewTokenWithFields(uuid.New(), userID, "", now.Add(time.Hour*24), now)

	repo := NewTokenRepository(testPool)
	if err := repo.SaveRefreshToken(context.Background(), *existingToken); err != nil {
		t.Fatalf("failed saving token for get test: %v", err)
	}

	type testCase struct {
		name      string
		tokenID   uuid.UUID
		wantErr   error
		wantFound bool
	}

	testCases := []testCase{
		{
			name:      "success get existing token",
			tokenID:   existingToken.ID(),
			wantErr:   nil,
			wantFound: true,
		},
		{
			name:      "failed token not found",
			tokenID:   uuid.New(),
			wantErr:   jwt.ErrNotFound,
			wantFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotToken, err := repo.GetRefreshTokenFromID(context.Background(), tc.tokenID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("GetRefreshTokenFromID() error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.wantFound {
				if gotToken == nil {
					t.Fatal("expected non-nil token")
				}
				if gotToken.ID() != existingToken.ID() {
					t.Errorf("ID = %v, want %v", gotToken.ID(), existingToken.ID())
				}
				if gotToken.UserID() != existingToken.UserID() {
					t.Errorf("UserID = %v, want %v", gotToken.UserID(), existingToken.UserID())
				}
				if !gotToken.ExpiresAt().Equal(existingToken.ExpiresAt()) {
					t.Errorf("ExpiresAt = %v, want %v", gotToken.ExpiresAt(), existingToken.ExpiresAt())
				}
				if !gotToken.IssuedAt().Equal(existingToken.IssuedAt()) {
					t.Errorf("IssuedAt = %v, want %v", gotToken.IssuedAt(), existingToken.IssuedAt())
				}
			}
		})
	}
}

func TestRepository_DeleteRefreshToken(t *testing.T) {
	truncateTables(t)

	userID := uuid.New()
	createTestUser(t, userID, "user_delete")

	now := time.Now().Truncate(time.Microsecond)
	tokenToDelete := auth2.NewTokenWithFields(uuid.New(), userID, "", now.Add(time.Hour), now)

	repo := NewTokenRepository(testPool)
	if err := repo.SaveRefreshToken(context.Background(), *tokenToDelete); err != nil {
		t.Fatalf("failed saving token: %v", err)
	}

	type testCase struct {
		name      string
		tokenID   uuid.UUID
		wantErr   error
		checkDead bool
	}

	testCases := []testCase{
		{
			name:      "success delete existing token",
			tokenID:   tokenToDelete.ID(),
			wantErr:   nil,
			checkDead: true,
		},
		{
			name:      "failed delete non-existent token",
			tokenID:   uuid.New(),
			wantErr:   jwt.ErrNotFound,
			checkDead: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.DeleteRefreshToken(context.Background(), tc.tokenID)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("DeleteRefreshToken() error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.checkDead {
				_, err := repo.GetRefreshTokenFromID(context.Background(), tc.tokenID)
				if !errors.Is(err, jwt.ErrNotFound) {
					t.Errorf("expected ErrNotFound after deletion, got %v", err)
				}
			}
		})
	}
}
