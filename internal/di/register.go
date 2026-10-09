package di

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gr0shka/Tic-tac-toe/internal/config"
	gameService "github.com/gr0shka/Tic-tac-toe/internal/domain/service"
	"github.com/gr0shka/Tic-tac-toe/internal/migration"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/postgres"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/postgres/game"
	userRepo "github.com/gr0shka/Tic-tac-toe/internal/repository/postgres/user"
	"github.com/gr0shka/Tic-tac-toe/internal/repository/transactor"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/handler"
	"github.com/gr0shka/Tic-tac-toe/internal/transport/http/middleware"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/app"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/auth"
	"github.com/gr0shka/Tic-tac-toe/internal/usecase/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			context.Background,
			config.Load,
			func(conf *config.Config) config.Postgres {
				return conf.Postgres
			},
			postgres.NewClient,
			fx.Annotate(
				game.New,
				fx.As(new(app.GameRepository)),
			),
			fx.Annotate(
				userRepo.New,
				fx.As(new(user.UserRepository)),
			),
			fx.Annotate(
				gameService.NewGameService,
				fx.As(new(gameService.GameService)),
			),
			fx.Annotate(
				transactor.NewTransactor,
				fx.As(new(usecase.Transactor)),
			),
			fx.Annotate(
				app.NewAppService,
				fx.As(new(app.AppService)),
			),
			fx.Annotate(
				user.NewUserService,
				fx.As(new(user.UserService)),
			),
			fx.Annotate(
				auth.NewAuthenticateService,
				fx.As(new(auth.AuthenticateService)),
			),
			middleware.NewUserAuthenticator,
			handler.NewGameHandler,
			handler.NewUserHandler,
			handler.NewAuthHandler,
		),
		fx.Invoke(
			RegisterServer,
		),
	)
}

func RegisterServer(lc fx.Lifecycle, gh *handler.GameHandler, uh *handler.UserHandler, ah *handler.AuthHandler, m *middleware.UserAuthenticator, cfg *config.Config, pool *pgxpool.Pool) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/register", ah.Register)
	mux.HandleFunc("POST /user/auth", ah.Authenticate)

	mux.Handle("GET /user/{uuid}", m.Authenticate(http.HandlerFunc(uh.GetUserByID)))

	mux.Handle("POST /game/{uuid}", m.Authenticate(http.HandlerFunc(gh.NextTurn)))
	mux.Handle("POST /games", m.Authenticate(http.HandlerFunc(gh.NewGame)))
	mux.Handle("GET /games/{uuid}", m.Authenticate(http.HandlerFunc(gh.GetGame)))
	mux.Handle("POST /games/{uuid}/join", m.Authenticate(http.HandlerFunc(gh.JoinGame)))
	mux.Handle("GET /games", m.Authenticate(http.HandlerFunc(gh.AllGames)))

	muxWithMiddleware := middleware.CORS(mux)
	muxWithMiddleware = middleware.SetHeaders(muxWithMiddleware)

	srv := &http.Server{
		Addr:         cfg.App.Host + ":" + cfg.App.Port,
		Handler:      muxWithMiddleware,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			if cfg.Postgres.Migration == "true" {
				if err := migration.Up(ctx, pool); err != nil {
					log.Printf("Error migrating database: %s", err)
				}
			}

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
