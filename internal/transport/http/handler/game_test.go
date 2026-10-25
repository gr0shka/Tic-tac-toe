package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
)

type mockAppService struct {
	cg      *game.CurrentGame
	sliceCG []*game.CurrentGame
	outErr  error
	winner  int
	status  game.GameStatus
	lb      *game.LeaderBoard
}

func (m mockAppService) GetLeaderBoard(ctx context.Context, count int) (*game.LeaderBoard, error) {
	return m.lb, m.outErr
}

func (m mockAppService) AllEndedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]*game.CurrentGame, error) {
	return m.sliceCG, m.outErr
}

func (m mockAppService) CreateGame(ctx context.Context, id uuid.UUID, mode game.GameMode) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) ProcessPlayerMove(ctx context.Context, playerID, gameID uuid.UUID, board [3][3]int) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, game.GameStatus, error) {
	return m.winner, m.status, m.outErr
}

func (m mockAppService) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) AllGames(ctx context.Context) ([]*game.CurrentGame, error) {
	return m.sliceCG, m.outErr
}

func (m mockAppService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func TestHandler_GetGame(t *testing.T) {
	type testCase struct {
		name         string
		mockUUID     string
		mockErr      error
		expectedCode int
	}

	newUUID := uuid.New()

	testCases := []testCase{
		{
			name:         "success get game",
			mockUUID:     newUUID.String(),
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid uuid",
			mockUUID:     "",
			expectedCode: http.StatusBadRequest,
			mockErr:      nil,
		},
		{
			name:         "game not found",
			mockUUID:     newUUID.String(),
			mockErr:      game.ErrNotFound,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/games/" + tc.mockUUID
			req, _ := http.NewRequest("GET", url, nil)
			req.SetPathValue("uuid", tc.mockUUID)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)

			mockGame := game.NewCurrentGame(gb)
			mockApp := mockAppService{cg: mockGame, outErr: tc.mockErr, status: game.StatusPlayerTurn}

			h := handler.NewGameHandler(mockApp)

			h.GetGame(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var actualResponse dto.GameBoardResponse
				err = json.Unmarshal(body, &actualResponse)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				cmpFieldGameBoardResponse(mockGame, t, actualResponse)
			}
		})
	}

}

func TestHandler_NewGame(t *testing.T) {
	type testCase struct {
		name           string
		mockBody       string
		mockErr        error
		mockPlayerUUID uuid.UUID
		expectedCode   int
	}

	newUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success new game",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusCreated,
			mockPlayerUUID: newUUID,
		},
		{
			name:           "invalid uuid",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusBadRequest,
			mockErr:        game.ErrInvalidArgument,
			mockPlayerUUID: uuid.Nil,
		},
		{
			name:           "empty body",
			mockBody:       "nil",
			expectedCode:   http.StatusBadRequest,
			mockErr:        nil,
			mockPlayerUUID: newUUID,
		},
		{
			name:           "error creating game",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusInternalServerError,
			mockErr:        errors.New("error creating game"),
			mockPlayerUUID: newUUID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/games"
			req, _ := http.NewRequest("POST", url, strings.NewReader(tc.mockBody))

			ctx := middleware.ContextWithUserID(context.Background(), tc.mockPlayerUUID)
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)

			mockGame := game.NewCurrentGame(gb)
			mockApp := mockAppService{cg: mockGame, outErr: tc.mockErr, status: game.StatusWaitingForPlayers}

			h := handler.NewGameHandler(mockApp)

			h.NewGame(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusCreated {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var actualResponse dto.GameBoardResponse
				err = json.Unmarshal(body, &actualResponse)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				cmpFieldGameBoardResponse(mockGame, t, actualResponse)
			}
		})
	}

}

func TestHandler_NextTurn(t *testing.T) {
	type testCase struct {
		name           string
		mockBody       string
		mockErr        error
		mockPlayerUUID uuid.UUID
		mockGameUUID   uuid.UUID
		expectedCode   int
	}

	newPlayerUUID := uuid.New()
	newGameUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success next turn",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusOK,
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   newGameUUID,
		},
		{
			name:           "invalid game uuid",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   uuid.Nil,
			mockErr:        game.ErrInvalidArgument,
		},
		{
			name:           "invalid player uuid",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: uuid.Nil,
			mockGameUUID:   newGameUUID,
			mockErr:        game.ErrInvalidArgument,
		},
		{
			name:           "invalid body",
			mockBody:       "",
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   newGameUUID,
		},
		{
			name:           "invalid turn",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   newGameUUID,
			mockErr:        game.ErrInvalidMove,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/game/" + tc.mockPlayerUUID.String()

			req, _ := http.NewRequest("POST", url, strings.NewReader(tc.mockBody))
			req.SetPathValue("uuid", tc.mockGameUUID.String())

			ctx := middleware.ContextWithUserID(context.Background(), tc.mockPlayerUUID)
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)
			player2 := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
			gb.AddPlayer(player2)

			mockGame := game.NewCurrentGame(gb)
			mockApp := mockAppService{cg: mockGame, outErr: tc.mockErr, status: game.StatusPlayerTurn}

			h := handler.NewGameHandler(mockApp)

			h.NextTurn(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var actualResponse dto.GameBoardResponse
				err = json.Unmarshal(body, &actualResponse)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				cmpFieldGameBoardResponse(mockGame, t, actualResponse)
			}
		})
	}
}

func TestHandler_JoinGame(t *testing.T) {
	type testCase struct {
		name           string
		mockPlayerUUID uuid.UUID
		mockGameUUID   uuid.UUID
		mockErr        error
		expectedCode   int
	}

	newPlayerUUID := uuid.New()
	newGameUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success join game",
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   newGameUUID,
			mockErr:        nil,
			expectedCode:   http.StatusOK,
		},
		{
			name:           "invalid user uuid",
			mockPlayerUUID: uuid.Nil,
			mockGameUUID:   newGameUUID,
			mockErr:        game.ErrInvalidArgument,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "invalid game uuid",
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   uuid.Nil,
			mockErr:        game.ErrInvalidArgument,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "invalid join game",
			mockPlayerUUID: newPlayerUUID,
			mockGameUUID:   newGameUUID,
			mockErr:        game.ErrGameFull,
			expectedCode:   http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/games/" + tc.mockGameUUID.String() + "/join"

			ctx := middleware.ContextWithUserID(context.Background(), tc.mockPlayerUUID)

			req, _ := http.NewRequest("POST", url, nil)
			req.SetPathValue("uuid", tc.mockGameUUID.String())
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)

			mockGame := game.NewCurrentGame(gb)
			mockGame.SetStatus(game.StatusWaitingForPlayers)
			mockApp := mockAppService{cg: mockGame, outErr: tc.mockErr, status: game.StatusWaitingForPlayers}

			h := handler.NewGameHandler(mockApp)

			h.JoinGame(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var actualResponse dto.GameBoardResponse
				err = json.Unmarshal(body, &actualResponse)
				if err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				cmpFieldGameBoardResponse(mockGame, t, actualResponse)
			}
		})
	}
}

func TestHandler_AllGames(t *testing.T) {
	type testCase struct {
		name         string
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success get all games",
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid get all games",
			mockErr:      game.ErrNotFound,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", "/games", nil)

			gb1 := game.NewGameBoard()
			gb2 := game.NewGameBoard()

			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			player2 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)

			gb1.AddPlayer(player1)
			gb2.AddPlayer(player2)

			mockGame1 := game.NewCurrentGame(gb1)
			mockGame2 := game.NewCurrentGame(gb2)

			mockGame1.SetStatus(game.StatusWaitingForPlayers)
			mockGame2.SetStatus(game.StatusWaitingForPlayers)

			mockCGSlice := make([]*game.CurrentGame, 0, 2)
			mockCGSlice = append(mockCGSlice, mockGame1)
			mockCGSlice = append(mockCGSlice, mockGame2)

			mockAS := mockAppService{sliceCG: mockCGSlice, outErr: tc.mockErr, status: game.StatusPlayerTurn}

			h := handler.NewGameHandler(mockAS)

			h.AllGames(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				var dtoSlice []dto.GameBoardResponse

				err := json.NewDecoder(res.Body).Decode(&dtoSlice)
				if err != nil {
					t.Errorf("could not parse response body")
				}

				for i, d := range dtoSlice {

					if i < len(mockCGSlice) {
						cmpFieldGameBoardResponse(mockCGSlice[i], t, d)
					}
				}
			}
		})
	}
}

func cmpFieldGameBoardResponse(
	mockGame *game.CurrentGame,
	t *testing.T,
	actualResponse dto.GameBoardResponse,
) {

	if mockGame.ID() != actualResponse.ID {
		t.Errorf("expected id %s, got %s", mockGame.ID(), actualResponse.ID)
	}

	if string(mockGame.Status()) != actualResponse.Status {
		t.Errorf("expected status %s, got %s", mockGame.Status(), actualResponse.Status)
	}

	if mockGame.ActivePlayer() != nil && mockGame.ActivePlayer().ID() != *actualResponse.ActivePlayer {
		t.Errorf("expected id %s, got %s", mockGame.ActivePlayer().ID(), *actualResponse.ActivePlayer)
	}

	if mockGame.Players()[game.FirstPlayer].Symbol() != actualResponse.Player1Symbol {
		t.Errorf("expected player1_symbol %v, got %v",
			mockGame.Players()[game.FirstPlayer].Symbol(), actualResponse.Player1Symbol)
	}

	if len(mockGame.Players()) == game.CountPlayers && mockGame.Players()[game.SecondPlayer] != nil {

		if mockGame.Players()[game.SecondPlayer].Symbol() != *actualResponse.Player2Symbol {
			t.Errorf("expected player2_symbol %v, got %v",
				mockGame.Players()[game.SecondPlayer].Symbol(),
				*actualResponse.Player2Symbol)
		}
	}

	if mockGame.Winner() != actualResponse.Winner {
		t.Errorf("expected winner %d, got %d", mockGame.Winner(), actualResponse.Winner)
	}

	if mockGame.GameBoard.Board() != actualResponse.Board {
		t.Errorf("expected board %d, got %d", mockGame.GameBoard.Board(), actualResponse.Board)
	}
}

func TestHandler_AllEndedGames(t *testing.T) {
	validUserID := uuid.New()

	gb := game.NewGameBoard()
	p1 := game.NewPlayer(validUserID, game.FirstPlayer, true)
	p2 := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
	_ = gb.AddPlayer(p1)
	_ = gb.AddPlayer(p2)

	cg := game.NewCurrentGame(gb)
	cg.SetStatus(game.StatusPlayerWins)
	cg.SetWinner(game.FirstPlayer)

	type testCase struct {
		name         string
		setAuthUser  bool
		mockGames    []*game.CurrentGame
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success get ended games",
			setAuthUser:  true,
			mockGames:    []*game.CurrentGame{cg},
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed user not in context",
			setAuthUser:  false,
			mockGames:    nil,
			mockErr:      nil,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "failed service error",
			setAuthUser:  true,
			mockGames:    nil,
			mockErr:      errors.New("db error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/games/ended", nil)

			if tc.setAuthUser {
				ctx := middleware.ContextWithUserID(req.Context(), validUserID)
				req = req.WithContext(ctx)
			}

			mockAS := mockAppService{
				sliceCG: tc.mockGames,
				outErr:  tc.mockErr,
			}

			h := handler.NewGameHandler(mockAS)
			h.AllEndedGames(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				var responseSlice []dto.GameBoardResponse
				if err := json.NewDecoder(res.Body).Decode(&responseSlice); err != nil {
					t.Fatalf("could not parse response body: %v", err)
				}
				if len(responseSlice) != len(tc.mockGames) {
					t.Errorf("expected %d games, got %d", len(tc.mockGames), len(responseSlice))
				}
				if len(responseSlice) > 0 {
					cmpFieldGameBoardResponse(tc.mockGames[0], t, responseSlice[0])
				}
			}
		})
	}
}

func TestHandler_GetLeaderBoard(t *testing.T) {
	playerID := uuid.New()
	p1, _ := game.NewPlayerWinRatio(playerID, "player1", 5, 2, 1, 0.625)
	lb := game.NewLeaderBoard(10)
	lb.Add(*p1)

	type testCase struct {
		name         string
		url          string
		mockLB       *game.LeaderBoard
		mockErr      error
		expectedCode int
	}

	testCases := []testCase{
		{
			name:         "success default limit",
			url:          "/leaderboard",
			mockLB:       lb,
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "success custom limit",
			url:          "/leaderboard?limit=25",
			mockLB:       lb,
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "success limit capped at 100",
			url:          "/leaderboard?limit=200",
			mockLB:       lb,
			mockErr:      nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed invalid limit parameter",
			url:          "/leaderboard?limit=invalid",
			mockLB:       nil,
			mockErr:      nil,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "failed service error",
			url:          "/leaderboard",
			mockLB:       nil,
			mockErr:      errors.New("db query error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)

			mockAS := mockAppService{
				lb:     tc.mockLB,
				outErr: tc.mockErr,
			}

			h := handler.NewGameHandler(mockAS)
			h.GetLeaderBoard(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Fatalf("expected status %d, got %d", tc.expectedCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				var response dto.LeaderBoardResponse
				if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
					t.Fatalf("could not parse response body: %v", err)
				}
				if len(response.Player) != tc.mockLB.Length() {
					t.Errorf("expected %d players, got %d", tc.mockLB.Length(), len(response.Player))
				}
				if len(response.Player) > 0 {
					p := response.Player[0]
					if p.ID != playerID {
						t.Errorf("expected ID %v, got %v", playerID, p.ID)
					}
					if p.Login != "player1" {
						t.Errorf("expected Login 'player1', got '%v'", p.Login)
					}
					if p.Wins != 5 {
						t.Errorf("expected Wins 5, got %d", p.Wins)
					}
					if p.Loss != 2 {
						t.Errorf("expected Loss 2, got %d", p.Loss)
					}
					if p.Draw != 1 {
						t.Errorf("expected Draw 1, got %d", p.Draw)
					}
					if p.Ratio != 0.625 {
						t.Errorf("expected Ratio 0.625, got %f", p.Ratio)
					}
				}
			}
		})
	}
}
