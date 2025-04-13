package main

import (
	"database/sql"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/llravell/go-pass/config"
	"github.com/llravell/go-pass/internal/grpc/server"
	"github.com/llravell/go-pass/internal/repository"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	"github.com/llravell/go-pass/logger"
	"github.com/llravell/go-pass/pkg/auth"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"google.golang.org/grpc"
)

func main() {
	log := logger.Get()

	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config building failed")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("open db error")
	}

	defer db.Close()

	listen, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Error().Err(err).Msg("tcp listen error")

		return
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	usersRepository := repository.NewUsersRepository(db)
	passwordsRepository := repository.NewPasswordsPostgresRepository(db)
	cardsRepository := repository.NewCardsPostgresRepository(db)

	authUsecase := usecase.NewAuthUseCase(usersRepository, jwtManager)
	passwordsUsecase := usecase.NewPasswordsUseCase(passwordsRepository)
	cardsUsecase := usecase.NewCardsUseCase(cardsRepository)

	authServer := server.NewAuthServer(authUsecase, &log)
	passwordsServer := server.NewPasswordsServer(passwordsUsecase, &log)
	cardsServer := server.NewCardsServer(cardsUsecase, &log)

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.AuthInterceptor(jwtManager),
			logging.UnaryServerInterceptor(server.Logger(&log), loggingOpts...),
		),
	)
	pb.RegisterAuthServer(srv, authServer)
	pb.RegisterPasswordsServer(srv, passwordsServer)
	pb.RegisterCardsServer(srv, cardsServer)

	log.Info().Msgf("server started on %s", cfg.Addr)

	if err := srv.Serve(listen); err != nil {
		log.Error().Err(err).Msg("server has been closed")
	}
}
