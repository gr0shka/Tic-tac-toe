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
}

func (m mockAppService) CreateGame(ctx context.Context, id uuid.UUID, mode game.GameMode) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) ProcessPlayerMove(ctx context.Context, playerID, gameID uuid.UUID, board [3][3]int) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, game.GameStatus) {
	return m.winner, m.status
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
			mockApp := mockAppService{mockGame, nil, tc.mockErr, 0, game.StatusPlayerTurn}

			h := handler.NewGameHandler(mockApp)

			h.GetGame(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedCode {
				t.Errorf("expected status 200 OK, got %d", res.StatusCode)
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
		mockPlayerUUID string
		expectedCode   int
	}

	newUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success new game",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusCreated,
			mockPlayerUUID: newUUID.String(),
		},
		{
			name:           "invalid uuid",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusBadRequest,
			mockErr:        nil,
			mockPlayerUUID: "",
		},
		{
			name:           "empty body",
			mockBody:       "nil",
			expectedCode:   http.StatusBadRequest,
			mockErr:        nil,
			mockPlayerUUID: newUUID.String(),
		},
		{
			name:           "error creating game",
			mockBody:       fmt.Sprintf(`{"mode":"%s"}`, game.GameModePlayerVSBot),
			expectedCode:   http.StatusBadRequest,
			mockErr:        errors.New("error creating game"),
			mockPlayerUUID: newUUID.String(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/games"
			req, _ := http.NewRequest("POST", url, strings.NewReader(tc.mockBody))

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey{}, tc.mockPlayerUUID)
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)

			mockGame := game.NewCurrentGame(gb)
			mockApp := mockAppService{mockGame, nil, tc.mockErr, 0, game.StatusWaitingForPlayers}

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
		mockPlayerUUID string
		mockGameUUID   string
		expectedCode   int
	}

	newPlayerUUID := uuid.New()
	newGameUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success next turn",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusOK,
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   newGameUUID.String(),
		},
		{
			name:           "invalid game uuid",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   "",
		},
		{
			name:           "invalid player uuid",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: "",
			mockGameUUID:   newGameUUID.String(),
		},
		{
			name:           "invalid body",
			mockBody:       "",
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   newGameUUID.String(),
		},
		{
			name:           "invalid turn",
			mockBody:       `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`,
			expectedCode:   http.StatusBadRequest,
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   newGameUUID.String(),
			mockErr:        game.ErrInvalidTurn,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/game/" + tc.mockPlayerUUID

			req, _ := http.NewRequest("POST", url, strings.NewReader(tc.mockBody))
			req.SetPathValue("uuid", tc.mockGameUUID)

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey{}, tc.mockPlayerUUID)
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)
			player2 := game.NewPlayer(uuid.New(), game.SecondPlayer, false)
			gb.AddPlayer(player2)

			mockGame := game.NewCurrentGame(gb)
			mockApp := mockAppService{mockGame, nil, tc.mockErr, 0, game.StatusPlayerTurn}

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
		mockPlayerUUID string
		mockGameUUID   string
		mockErr        error
		expectedCode   int
	}

	newPlayerUUID := uuid.New()
	newGameUUID := uuid.New()

	testCases := []testCase{
		{
			name:           "success join game",
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   newGameUUID.String(),
			mockErr:        nil,
			expectedCode:   http.StatusOK,
		},
		{
			name:           "invalid user uuid",
			mockPlayerUUID: "",
			mockGameUUID:   newGameUUID.String(),
			mockErr:        nil,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "invalid game uuid",
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   "",
			mockErr:        nil,
			expectedCode:   http.StatusBadRequest,
		},
		{
			name:           "invalid join game",
			mockPlayerUUID: newPlayerUUID.String(),
			mockGameUUID:   newGameUUID.String(),
			mockErr:        game.ErrMaxCountOfPlayers,
			expectedCode:   http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()

			url := "/games/" + tc.mockGameUUID + "/join"

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey{}, tc.mockPlayerUUID)

			req, _ := http.NewRequest("POST", url, nil)
			req.SetPathValue("uuid", tc.mockGameUUID)
			req = req.WithContext(ctx)

			gb := game.NewGameBoard()
			player1 := game.NewPlayer(uuid.New(), game.FirstPlayer, true)
			gb.AddPlayer(player1)

			mockGame := game.NewCurrentGame(gb)
			mockGame.SetStatus(game.StatusWaitingForPlayers)
			mockApp := mockAppService{mockGame, nil, tc.mockErr, 0, game.StatusWaitingForPlayers}

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

			mockAS := mockAppService{nil, mockCGSlice, tc.mockErr, 0, game.StatusPlayerTurn}

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
