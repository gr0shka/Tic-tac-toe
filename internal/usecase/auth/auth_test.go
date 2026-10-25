package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	auth2 "github.com/gr0shka/Tic-tac-toe/internal/domain/auth"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/user"
	"github.com/gr0shka/Tic-tac-toe/internal/security/jwt"
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

type mockJWTService struct {
	token *auth2.Token
	err   error
}

func (m mockJWTService) GenerateAccessToken(u user.User) (*auth2.Token, error) {
	return &auth2.Token{}, nil
}

func (m mockJWTService) GenerateRefreshToken(u user.User) (*auth2.Token, error) {
	return &auth2.Token{}, nil
}

func (m mockJWTService) ValidateAccessToken(at string) (*auth2.Token, error) {
	return m.token, m.err
}

func (m mockJWTService) ValidateRefreshToken(rt string) (*auth2.Token, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockJWTService) GetUserIDFromToken(t string) (uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockJWTService) GetTokenIDFromToken(t string) (uuid.UUID, error) {
	//TODO implement me
	panic("implement me")
}

type mockRepo struct {
}

func (m mockRepo) SaveRefreshToken(ctx context.Context, t auth2.Token) error {
	return nil
}

func (m mockRepo) GetRefreshTokenFromID(ctx context.Context, id uuid.UUID) (*auth2.Token, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockRepo) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

type mockTransactor struct {
}

func (m mockTransactor) Do(ctx context.Context, fn func(txCtx context.Context) error) error {
	//TODO implement me
	panic("implement me")
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
			err:      user.ErrInvalidArgument,
		},
		{
			name:     "Error authentication invalid password",
			login:    "testLogin",
			password: "",
			user:     testUser,
			err:      user.ErrInvalidArgument,
		},
		{
			name:     "Error authentication password not match",
			login:    "testLogin",
			password: "testPasswordErr",
			user:     testUser,
			err:      user.ErrInvalidCredentials,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			us := mockUserService{tc.user, tc.mockErr}
			pr := mockJWTService{}
			repo := mockRepo{}
			tx := mockTransactor{}
			as := auth.NewAuthenticateService(us, pr, repo, tx)

			_, err := as.Authenticate(context.Background(), tc.login, tc.password)

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
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
			err:      user.ErrInvalidArgument,
			mockErr:  nil,
		},
		{
			name:     "Error registration invalid password",
			login:    "testLogin1",
			password: "",
			user:     testUser,
			err:      user.ErrInvalidArgument,
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
			pr := mockJWTService{}
			repo := mockRepo{}
			tx := mockTransactor{}
			as := auth.NewAuthenticateService(us, pr, repo, tx)

			req := auth.SignUpRequest{tc.login, tc.password}
			err := as.Register(context.Background(), req)

			if !errors.Is(err, tc.err) {
				t.Errorf("error = %v, want = %v", err, tc.err)
			}
		})
	}
}

func TestAuthenticateService_AuthenticateByToken(t *testing.T) {
	validUserID := uuid.New()
	validateErr := errors.New("invalid signature")

	type testCase struct {
		name      string
		tokenStr  string
		mockToken *auth2.Token
		mockErr   error
		wantID    uuid.UUID
		wantErr   error
	}

	testCases := []testCase{
		{
			name:      "Success authenticate by token",
			tokenStr:  "valid.jwt.token",
			mockToken: auth2.NewToken(validUserID, time.Hour),
			mockErr:   nil,
			wantID:    validUserID,
			wantErr:   nil,
		},
		{
			name:      "Error validate access token",
			tokenStr:  "corrupted.jwt.token",
			mockToken: nil,
			mockErr:   validateErr,
			wantID:    uuid.Nil,
			wantErr:   validateErr,
		},
		{
			name:      "Error token has nil user id",
			tokenStr:  "token.with.nil.uid",
			mockToken: auth2.NewToken(uuid.Nil, time.Hour),
			mockErr:   nil,
			wantID:    uuid.Nil,
			wantErr:   jwt.ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := mockJWTService{
				token: tc.mockToken,
				err:   tc.mockErr,
			}
			as := auth.NewAuthenticateService(mockUserService{}, pr, mockRepo{}, mockTransactor{})

			gotID, err := as.AuthenticateByToken(tc.tokenStr)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("AuthenticateByToken() error = %v, wantErr %v", err, tc.wantErr)
			}

			if gotID != tc.wantID {
				t.Errorf("AuthenticateByToken() gotID = %v, wantID %v", gotID, tc.wantID)
			}
		})
	}
}
