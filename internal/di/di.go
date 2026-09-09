package di

import (
	"net/http"

	appService "github.com/gr0shka/Tic-tac-toe/internal/application/service"
	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapstore"
	transportHttp "github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"go.uber.org/fx"
)

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				mapstore.NewMapRepository,
				fx.As(new(repository.Repository)),
			),
			fx.Annotate(
				gameService.NewGameService,
				fx.As(new(gameService.GameService)),
			),
			fx.Annotate(
				appService.NewAppService,
				fx.As(new(appService.AppService)),
			),
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
