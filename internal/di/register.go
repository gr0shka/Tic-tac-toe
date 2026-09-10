package di

import (
	"net/http"

	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/repository"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/mapstore"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
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
				usecase.NewAppService,
				fx.As(new(usecase.AppService)),
			),
			handler.NewHandler,
		),
		fx.Invoke(
			CreateMuxAndStartServer,
		),
	)
}

func CreateMuxAndStartServer(h *handler.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /game/{uuid}", h.NextTurn)
	mux.HandleFunc("GET /create", h.NewGame)
	mux.HandleFunc("GET /get/{uuid}", h.GetGame)

	muxWithMiddleware := middleware.MiddlewareCorsResponse(mux)
	muxWithMiddleware = middleware.MiddlewareSetHeaders(muxWithMiddleware)

	go func() {
		http.ListenAndServe(":8080", muxWithMiddleware)
	}()
}
