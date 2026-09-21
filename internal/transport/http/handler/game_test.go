package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
)

type mockAppService struct {
	cg     *game.CurrentGame
	outErr error
	winner int
	isEnd  bool
}

func (m mockAppService) CreateGame(ctx context.Context) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) ProcessPlayerMove(ctx context.Context, id uuid.UUID, board [3][3]int) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) GetGame(ctx context.Context, id uuid.UUID) (*game.CurrentGame, error) {
	return m.cg, m.outErr
}

func (m mockAppService) GameIsEnded(ctx context.Context, id uuid.UUID) (int, bool) {
	return m.winner, m.isEnd
}

func TestHandler_GetGame_Success(t *testing.T) {
	w := httptest.NewRecorder()

	gameID := uuid.New()
	url := "/games/" + gameID.String()
	req, _ := http.NewRequest("GET", url, nil)
	req.SetPathValue("uuid", gameID.String())

	gb := game.NewGameBoard()
	mockGame := game.NewCurrentGame(gb)
	mockApp := mockAppService{mockGame, nil, 0, false}

	h := handler.NewGameHandler(mockApp)

	h.GetGame(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var actualResponse dto.GameBoardResponse
	err = json.Unmarshal(body, &actualResponse)
	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
}

func TestHandler_GetGame_NotFound(t *testing.T) {
	w := httptest.NewRecorder()

	gameID := uuid.New()
	url := "/games/" + gameID.String()
	req, _ := http.NewRequest("GET", url, nil)
	req.SetPathValue("uuid", gameID.String())

	gb := game.NewGameBoard()
	mockGame := game.NewCurrentGame(gb)
	mockApp := mockAppService{mockGame, game.ErrNotFound, 0, false}

	h := handler.NewGameHandler(mockApp)

	h.GetGame(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

}

func TestHandler_NewGame(t *testing.T) {
	w := httptest.NewRecorder()

	url := "/games"
	req, _ := http.NewRequest("POST", url, nil)

	gb := game.NewGameBoard()
	mockGame := game.NewCurrentGame(gb)
	mockApp := mockAppService{mockGame, nil, 0, false}

	h := handler.NewGameHandler(mockApp)

	h.NewGame(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var actualResponse dto.GameBoardResponse
	err = json.Unmarshal(body, &actualResponse)
	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
}

func TestHandler_NextTurn(t *testing.T) {
	w := httptest.NewRecorder()

	gameID := uuid.New()
	url := "/game/" + gameID.String()

	jsonBody := `{"board": [[1,0,0],[0,2,0],[0,0,0]]}`

	req, _ := http.NewRequest("POST", url, strings.NewReader(jsonBody))
	req.SetPathValue("uuid", gameID.String())

	gb := game.NewGameBoard()
	mockGame := game.NewCurrentGame(gb)
	mockApp := mockAppService{mockGame, nil, 0, false}

	h := handler.NewGameHandler(mockApp)

	h.NextTurn(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var actualResponse dto.GameBoardResponse
	err = json.Unmarshal(body, &actualResponse)
	if err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
}
