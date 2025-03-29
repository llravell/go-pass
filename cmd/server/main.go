package main

import (
	"database/sql"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	cfg, err := buildConfig()

	if err != nil {
		log.Fatal().Err(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("open db error")
	}

	defer db.Close()

	listen, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("tcp listen error")
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	usersRepository := repository.NewUsersRepository(db)
	authUsecase := usecase.NewAuthUseCase(usersRepository, jwtManager)

	authServer := server.NewAuthServer(authUsecase, &log)

	server := grpc.NewServer()
	pb.RegisterAuthServer(server, authServer)

	if err := server.Serve(listen); err != nil {
		log.Error().Err(err).Msg("server has been closed")
	}
}
