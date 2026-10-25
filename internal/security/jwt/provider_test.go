package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/config"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
)

func writeRSAKeysToDir(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed generating rsa key: %v", err)
	}

	privASN1, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("failed marshalling private key: %v", err)
	}

	privBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privASN1,
	}

	privPath = filepath.Join(dir, "private.pem")
	privFile, err := os.Create(privPath)
	if err != nil {
		t.Fatalf("failed creating priv file: %v", err)
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, privBlock); err != nil {
		t.Fatalf("failed encoding private pem: %v", err)
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		t.Fatalf("failed marshalling public key: %v", err)
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}

	pubPath = filepath.Join(dir, "public.pem")
	pubFile, err := os.Create(pubPath)
	if err != nil {
		t.Fatalf("failed creating pub file: %v", err)
	}
	defer pubFile.Close()
	if err := pem.Encode(pubFile, pubBlock); err != nil {
		t.Fatalf("failed encoding public pem: %v", err)
	}

	return privPath, pubPath
}

func TestNewProvider(t *testing.T) {
	tempDir := t.TempDir()
	privPath, pubPath := writeRSAKeysToDir(t, tempDir)

	badFile := filepath.Join(tempDir, "bad.pem")
	_ = os.WriteFile(badFile, []byte("invalid pem data"), 0644)

	type testCase struct {
		name    string
		cfg     config.JWT
		wantErr bool
	}

	testCases := []testCase{
		{
			name: "success create provider",
			cfg: config.JWT{
				PrivateFile: privPath,
				PublicFile:  pubPath,
				AccessTTL:   time.Minute * 15,
				RefreshTTL:  time.Hour * 24,
			},
			wantErr: false,
		},
		{
			name: "failed private key not found",
			cfg: config.JWT{
				PrivateFile: filepath.Join(tempDir, "missing_private.pem"),
				PublicFile:  pubPath,
			},
			wantErr: true,
		},
		{
			name: "failed public key not found",
			cfg: config.JWT{
				PrivateFile: privPath,
				PublicFile:  filepath.Join(tempDir, "missing_public.pem"),
			},
			wantErr: true,
		},
		{
			name: "failed invalid private pem content",
			cfg: config.JWT{
				PrivateFile: badFile,
				PublicFile:  pubPath,
			},
			wantErr: true,
		},
		{
			name: "failed invalid public pem content",
			cfg: config.JWT{
				PrivateFile: privPath,
				PublicFile:  badFile,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := jwt.NewProvider(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewProvider() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if !tc.wantErr && p == nil {
				t.Fatal("NewProvider() returned nil provider without error")
			}
		})
	}
}

func TestProvider_ValidateAccessToken(t *testing.T) {
	tempDir := t.TempDir()
	privPath, pubPath := writeRSAKeysToDir(t, tempDir)

	p, err := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   time.Hour,
		RefreshTTL:  time.Hour * 24,
	})
	if err != nil {
		t.Fatalf("failed creating provider: %v", err)
	}

	testUser := *user.New(uuid.New(), "alice", "hash")

	validTok, err := p.GenerateAccessToken(testUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	expiredP, _ := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   -time.Minute,
	})
	expiredTok, _ := expiredP.GenerateAccessToken(testUser)

	refreshTok, _ := p.GenerateRefreshToken(testUser)

	otherDir := t.TempDir()
	otherPriv, otherPub := writeRSAKeysToDir(t, otherDir)
	otherP, _ := jwt.NewProvider(config.JWT{
		PrivateFile: otherPriv,
		PublicFile:  otherPub,
		AccessTTL:   time.Hour,
	})
	otherKeyTok, _ := otherP.GenerateAccessToken(testUser)

	type testCase struct {
		name       string
		tokenStr   string
		wantErr    error
		wantUserID uuid.UUID
		wantTokID  uuid.UUID
	}

	testCases := []testCase{
		{
			name:       "success validate access token",
			tokenStr:   validTok.TokenString(),
			wantErr:    nil,
			wantUserID: testUser.ID(),
			wantTokID:  validTok.ID(),
		},
		{
			name:     "failed validate expired access token",
			tokenStr: expiredTok.TokenString(),
			wantErr:  jwt.ErrTokenExpired,
		},
		{
			name:     "failed validate token with wrong token type (refresh)",
			tokenStr: refreshTok.TokenString(),
			wantErr:  jwt.ErrInvalidToken,
		},
		{
			name:     "failed validate corrupted token",
			tokenStr: "corrupted.jwt.string",
			wantErr:  jwt.ErrInvalidToken,
		},
		{
			name:     "failed validate token signed with different key",
			tokenStr: otherKeyTok.TokenString(),
			wantErr:  jwt.ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := p.ValidateAccessToken(tc.tokenStr)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateAccessToken() error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if tok == nil {
					t.Fatal("expected non-nil token")
				}
				if tok.UserID() != tc.wantUserID {
					t.Errorf("UserID = %v, want %v", tok.UserID(), tc.wantUserID)
				}
				if tok.ID() != tc.wantTokID {
					t.Errorf("TokenID = %v, want %v", tok.ID(), tc.wantTokID)
				}
			}
		})
	}
}

func TestProvider_ValidateRefreshToken(t *testing.T) {
	tempDir := t.TempDir()
	privPath, pubPath := writeRSAKeysToDir(t, tempDir)

	p, err := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   time.Hour,
		RefreshTTL:  time.Hour * 24,
	})
	if err != nil {
		t.Fatalf("failed creating provider: %v", err)
	}

	testUser := *user.New(uuid.New(), "bob", "hash")

	validTok, err := p.GenerateRefreshToken(testUser)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	expiredP, _ := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   time.Hour,
		RefreshTTL:  -time.Minute,
	})
	expiredTok, _ := expiredP.GenerateRefreshToken(testUser)

	accessTok, _ := p.GenerateAccessToken(testUser)

	type testCase struct {
		name       string
		tokenStr   string
		wantErr    error
		wantUserID uuid.UUID
		wantTokID  uuid.UUID
	}

	testCases := []testCase{
		{
			name:       "success validate refresh token",
			tokenStr:   validTok.TokenString(),
			wantErr:    nil,
			wantUserID: testUser.ID(),
			wantTokID:  validTok.ID(),
		},
		{
			name:     "failed validate expired refresh token",
			tokenStr: expiredTok.TokenString(),
			wantErr:  jwt.ErrTokenExpired,
		},
		{
			name:     "failed validate refresh token with wrong token type (access)",
			tokenStr: accessTok.TokenString(),
			wantErr:  jwt.ErrInvalidToken,
		},
		{
			name:     "failed validate corrupted refresh token",
			tokenStr: "not-a-valid-token",
			wantErr:  jwt.ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := p.ValidateRefreshToken(tc.tokenStr)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateRefreshToken() error = %v, wantErr = %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if tok == nil {
					t.Fatal("expected non-nil token")
				}
				if tok.UserID() != tc.wantUserID {
					t.Errorf("UserID = %v, want %v", tok.UserID(), tc.wantUserID)
				}
				if tok.ID() != tc.wantTokID {
					t.Errorf("TokenID = %v, want %v", tok.ID(), tc.wantTokID)
				}
			}
		})
	}
}

func TestProvider_GetIDsFromToken(t *testing.T) {
	tempDir := t.TempDir()
	privPath, pubPath := writeRSAKeysToDir(t, tempDir)

	p, err := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   time.Hour,
		RefreshTTL:  time.Hour * 24,
	})
	if err != nil {
		t.Fatalf("failed creating provider: %v", err)
	}

	expectedUserID := uuid.New()
	testUser := *user.New(expectedUserID, "charlie", "hash")

	validTok, err := p.GenerateAccessToken(testUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	expiredP, _ := jwt.NewProvider(config.JWT{
		PrivateFile: privPath,
		PublicFile:  pubPath,
		AccessTTL:   -time.Minute,
	})
	expiredTok, _ := expiredP.GenerateAccessToken(testUser)

	t.Run("GetUserIDFromToken table test", func(t *testing.T) {
		type testCase struct {
			name     string
			tokenStr string
			wantID   uuid.UUID
			wantErr  error
		}

		testCases := []testCase{
			{
				name:     "success get user id",
				tokenStr: validTok.TokenString(),
				wantID:   expectedUserID,
				wantErr:  nil,
			},
			{
				name:     "failed corrupted token",
				tokenStr: "corrupted.token",
				wantID:   uuid.Nil,
				wantErr:  jwt.ErrInvalidToken,
			},
			{
				name:     "failed expired token",
				tokenStr: expiredTok.TokenString(),
				wantID:   uuid.Nil,
				wantErr:  jwt.ErrTokenExpired,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				uid, err := p.GetUserIDFromToken(tc.tokenStr)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("GetUserIDFromToken() error = %v, wantErr = %v", err, tc.wantErr)
				}
				if uid != tc.wantID {
					t.Errorf("GetUserIDFromToken() got %v, want %v", uid, tc.wantID)
				}
			})
		}
	})

	t.Run("GetTokenIDFromToken table test", func(t *testing.T) {
		type testCase struct {
			name     string
			tokenStr string
			wantID   uuid.UUID
			wantErr  error
		}

		testCases := []testCase{
			{
				name:     "success get token id",
				tokenStr: validTok.TokenString(),
				wantID:   validTok.ID(),
				wantErr:  nil,
			},
			{
				name:     "failed corrupted token",
				tokenStr: "corrupted.token",
				wantID:   uuid.Nil,
				wantErr:  jwt.ErrInvalidToken,
			},
			{
				name:     "failed expired token",
				tokenStr: expiredTok.TokenString(),
				wantID:   uuid.Nil,
				wantErr:  jwt.ErrTokenExpired,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tid, err := p.GetTokenIDFromToken(tc.tokenStr)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("GetTokenIDFromToken() error = %v, wantErr = %v", err, tc.wantErr)
				}
				if tid != tc.wantID {
					t.Errorf("GetTokenIDFromToken() got %v, want %v", tid, tc.wantID)
				}
			})
		}
	})
}
