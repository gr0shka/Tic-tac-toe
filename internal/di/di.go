package di

import (
	"net/http"

	appService "github.com/gr0shka/Tic-tac-toe/internal/application/service"
	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapstore"
	transportHttp "github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"go.uber.org/fx"
)

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			mapstore.NewMapRepository,
			gameService.NewGameService,
			appService.NewAppService,
			transportHttp.NewHandler,
		),
		fx.Invoke(
			CreateMuxAndStartServer,
		),
	)
}

func CreateMuxAndStartServer(h *transportHttp.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /game/{uuid}", h.NextTurn)
	mux.HandleFunc("GET /create", h.NewGame)
	mux.HandleFunc("GET /get/{uuid}", h.GetGame)

	go func() {
		http.ListenAndServe(":8080", mux)
	}()
}
