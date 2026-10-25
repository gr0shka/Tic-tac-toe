package mapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gr0shka/Tic-tac-toe/internal/domain/game"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/mapper"
)

func TestToGameBoardResponse(t *testing.T) {
	p1ID := uuid.New()
	p2ID := uuid.New()
	fixedTime := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

	createGame := func(
		twoPlayers bool,
		activeP *game.Player,
		status game.GameStatus,
		winner int,
	) *game.CurrentGame {
		gb := game.NewGameBoard()
		p1 := game.NewPlayer(p1ID, game.FirstPlayer, true)
		_ = gb.AddPlayer(p1)

		if twoPlayers {
			p2 := game.NewPlayer(p2ID, game.SecondPlayer, false)
			_ = gb.AddPlayer(p2)
		}

		cg := game.NewCurrentGame(gb)
		cg.SetActivePlayer(activeP)
		cg.SetStatus(status)
		cg.SetWinner(winner)
		cg.SetCreatedAt(fixedTime)
		return cg
	}

	p1 := game.NewPlayer(p1ID, game.FirstPlayer, true)
	p2 := game.NewPlayer(p2ID, game.SecondPlayer, false)

	type testCase struct {
		name              string
		cg                *game.CurrentGame
		wantPlayer1ID     uuid.UUID
		wantPlayer1Symbol int
		wantHasPlayer2    bool
		wantPlayer2ID     uuid.UUID
		wantPlayer2Symbol int
		wantHasActive     bool
		wantActiveID      uuid.UUID
		wantWinner        int
		wantHasWinnerID   bool
		wantWinnerID      uuid.UUID
		wantStatus        string
	}

	testCases := []testCase{
		{
			name:              "one player waiting for second player",
			cg:                createGame(false, nil, game.StatusWaitingForPlayers, -1),
			wantPlayer1ID:     p1ID,
			wantPlayer1Symbol: game.FirstPlayer,
			wantHasPlayer2:    false,
			wantHasActive:     false,
			wantWinner:        -1,
			wantHasWinnerID:   false,
			wantStatus:        string(game.StatusWaitingForPlayers),
		},
		{
			name:              "two players game in progress active player 2",
			cg:                createGame(true, p2, game.StatusPlayerTurn, -1),
			wantPlayer1ID:     p1ID,
			wantPlayer1Symbol: game.FirstPlayer,
			wantHasPlayer2:    true,
			wantPlayer2ID:     p2ID,
			wantPlayer2Symbol: game.SecondPlayer,
			wantHasActive:     true,
			wantActiveID:      p2ID,
			wantWinner:        -1,
			wantHasWinnerID:   false,
			wantStatus:        string(game.StatusPlayerTurn),
		},
		{
			name:              "game finished player 1 wins",
			cg:                createGame(true, p1, game.StatusPlayerWins, game.FirstPlayer),
			wantPlayer1ID:     p1ID,
			wantPlayer1Symbol: game.FirstPlayer,
			wantHasPlayer2:    true,
			wantPlayer2ID:     p2ID,
			wantPlayer2Symbol: game.SecondPlayer,
			wantHasActive:     true,
			wantActiveID:      p1ID,
			wantWinner:        game.FirstPlayer,
			wantHasWinnerID:   true,
			wantWinnerID:      p1ID,
			wantStatus:        string(game.StatusPlayerWins),
		},
		{
			name:              "game finished player 2 wins",
			cg:                createGame(true, p2, game.StatusPlayerWins, game.SecondPlayer),
			wantPlayer1ID:     p1ID,
			wantPlayer1Symbol: game.FirstPlayer,
			wantHasPlayer2:    true,
			wantPlayer2ID:     p2ID,
			wantPlayer2Symbol: game.SecondPlayer,
			wantHasActive:     true,
			wantActiveID:      p2ID,
			wantWinner:        game.SecondPlayer,
			wantHasWinnerID:   true,
			wantWinnerID:      p2ID,
			wantStatus:        string(game.StatusPlayerWins),
		},
		{
			name:              "game finished draw",
			cg:                createGame(true, nil, game.StatusDraw, -1),
			wantPlayer1ID:     p1ID,
			wantPlayer1Symbol: game.FirstPlayer,
			wantHasPlayer2:    true,
			wantPlayer2ID:     p2ID,
			wantPlayer2Symbol: game.SecondPlayer,
			wantHasActive:     false,
			wantWinner:        -1,
			wantHasWinnerID:   false,
			wantStatus:        string(game.StatusDraw),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := mapper.ToGameBoardResponse(tc.cg)

			if resp.ID != tc.cg.ID() {
				t.Errorf("ID = %v, want %v", resp.ID, tc.cg.ID())
			}
			if resp.Status != tc.wantStatus {
				t.Errorf("Status = %v, want %v", resp.Status, tc.wantStatus)
			}
			if resp.CreatedAt != fixedTime {
				t.Errorf("CreatedAt = %v, want %v", resp.CreatedAt, fixedTime)
			}
			if resp.Player1ID != tc.wantPlayer1ID {
				t.Errorf("Player1ID = %v, want %v", resp.Player1ID, tc.wantPlayer1ID)
			}
			if resp.Player1Symbol != tc.wantPlayer1Symbol {
				t.Errorf("Player1Symbol = %v, want %v", resp.Player1Symbol, tc.wantPlayer1Symbol)
			}

			// Player 2
			if tc.wantHasPlayer2 {
				if resp.Player2ID == nil || *resp.Player2ID != tc.wantPlayer2ID {
					t.Errorf("Player2ID = %v, want %v", resp.Player2ID, tc.wantPlayer2ID)
				}
				if resp.Player2Symbol == nil || *resp.Player2Symbol != tc.wantPlayer2Symbol {
					t.Errorf("Player2Symbol = %v, want %v", resp.Player2Symbol, tc.wantPlayer2Symbol)
				}
			} else {
				if resp.Player2ID != nil {
					t.Errorf("expected Player2ID nil, got %v", resp.Player2ID)
				}
				if resp.Player2Symbol != nil {
					t.Errorf("expected Player2Symbol nil, got %v", resp.Player2Symbol)
				}
			}

			// Active player
			if tc.wantHasActive {
				if resp.ActivePlayer == nil || *resp.ActivePlayer != tc.wantActiveID {
					t.Errorf("ActivePlayer = %v, want %v", resp.ActivePlayer, tc.wantActiveID)
				}
			} else {
				if resp.ActivePlayer != nil {
					t.Errorf("expected ActivePlayer nil, got %v", resp.ActivePlayer)
				}
			}

			// Winner
			if resp.Winner != tc.wantWinner {
				t.Errorf("Winner = %v, want %v", resp.Winner, tc.wantWinner)
			}
			if tc.wantHasWinnerID {
				if resp.WinnerID == nil || *resp.WinnerID != tc.wantWinnerID {
					t.Errorf("WinnerID = %v, want %v", resp.WinnerID, tc.wantWinnerID)
				}
			} else {
				if resp.WinnerID != nil {
					t.Errorf("expected WinnerID nil, got %v", resp.WinnerID)
				}
			}
		})
	}
}

func TestToLeaderBoardResponse(t *testing.T) {
	p1ID := uuid.New()
	p2ID := uuid.New()

	p1, _ := game.NewPlayerWinRatio(p1ID, "alice", 10, 2, 1, 0.8)
	p2, _ := game.NewPlayerWinRatio(p2ID, "bob", 5, 5, 0, 0.5)

	lbWithPlayers := game.NewLeaderBoard(2)
	lbWithPlayers.Add(*p1, *p2)

	type testCase struct {
		name      string
		lb        *game.LeaderBoard
		wantCount int
	}

	testCases := []testCase{
		{
			name:      "empty leaderboard",
			lb:        game.NewLeaderBoard(0),
			wantCount: 0,
		},
		{
			name:      "leaderboard with players",
			lb:        lbWithPlayers,
			wantCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := mapper.ToLeaderBoardResponse(tc.lb)

			if len(res.Player) != tc.wantCount {
				t.Fatalf("expected %d players, got %d", tc.wantCount, len(res.Player))
			}

			if tc.wantCount > 0 {
				first := res.Player[0]
				if first.ID != p1ID {
					t.Errorf("ID = %v, want %v", first.ID, p1ID)
				}
				if first.Login != "alice" {
					t.Errorf("Login = %s, want alice", first.Login)
				}
				if first.Wins != 10 {
					t.Errorf("Wins = %d, want 10", first.Wins)
				}
				if first.Loss != 2 {
					t.Errorf("Loss = %d, want 2", first.Loss)
				}
				if first.Draw != 1 {
					t.Errorf("Draw = %d, want 1", first.Draw)
				}
				if first.Ratio != 0.8 {
					t.Errorf("Ratio = %f, want 0.8", first.Ratio)
				}
			}
		})
	}
}
