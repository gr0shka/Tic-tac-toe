package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapstore"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/dto"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
)

func TestHandler_GetGame(t *testing.T) {
	w := httptest.NewRecorder()

	url := "/get/game/" + uuid.UUID{}.String()
	req, _ := http.NewRequest("GET", url, nil)

	gs := service.NewGameService()
	repo := mapstore.NewMapRepository()
	as := usecase.NewAppService(gs, repo)

	h := handler.NewHandler(as)

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

func TestHandler_NewGame(t *testing.T) {
	w := httptest.NewRecorder()

	url := "/create"
	req, _ := http.NewRequest("GET", url, nil)

	gs := service.NewGameService()
	repo := mapstore.NewMapRepository()
	as := usecase.NewAppService(gs, repo)

	h := handler.NewHandler(as)

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

	url := "/create"
	req, _ := http.NewRequest("GET", url, nil)

	gs := service.NewGameService()
	repo := mapstore.NewMapRepository()
	as := usecase.NewAppService(gs, repo)

	h := handler.NewHandler(as)

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
