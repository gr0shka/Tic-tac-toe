package di

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
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
				fx.As(new(usecase.Repository)),
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
			RegisterServer,
		),
	)
}

func RegisterServer(lc fx.Lifecycle, h *handler.Handler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /game/{uuid}", h.NextTurn)
	mux.HandleFunc("GET /create", h.NewGame)
	mux.HandleFunc("GET /get/{uuid}", h.GetGame)

	muxWithMiddleware := middleware.CORS(mux)
	muxWithMiddleware = middleware.SetHeaders(muxWithMiddleware)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      muxWithMiddleware,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("HTTP server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
