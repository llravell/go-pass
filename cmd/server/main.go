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
	"github.com/llravell/go-pass/pkg/workerpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
)

const (
	fileDeletingWorkersAmount = 4
)

//nolint:funlen
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

	minioClient, err := minio.New(cfg.MinioAddr, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.MinioAccessKeyID, cfg.MinioSecretAccessKey, ""),
	})
	if err != nil {
		log.Error().Err(err).Msg("initialize minio client error")

		return
	}

	listen, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Error().Err(err).Msg("tcp listen error")

		return
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	usersRepository := repository.NewUsersRepository(db)
	passwordsRepository := repository.NewPasswordsPostgresRepository(db)
	cardsRepository := repository.NewCardsPostgresRepository(db)
	filesRepository := repository.NewFilesPostgresRepository(db)
	filesS3Storage := repository.NewFilesMinioStorage(minioClient)

	fileDeletingWorkerPool := workerpool.New[*usecase.FileDeleteWork](fileDeletingWorkersAmount)

	authUsecase := usecase.NewAuthUseCase(usersRepository, jwtManager)
	passwordsUsecase := usecase.NewPasswordsUseCase(passwordsRepository)
	cardsUsecase := usecase.NewCardsUseCase(cardsRepository)
	filesUsecase := usecase.NewFilesUseCase(filesRepository, filesS3Storage, fileDeletingWorkerPool, &log)

	authServer := server.NewAuthServer(authUsecase, &log)
	passwordsServer := server.NewPasswordsServer(passwordsUsecase, &log)
	cardsServer := server.NewCardsServer(cardsUsecase, &log)
	notesServer := server.NewNotesServer(filesUsecase, &log)

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.AuthInterceptor(jwtManager),
			logging.UnaryServerInterceptor(server.Logger(&log), loggingOpts...),
		),
		grpc.ChainStreamInterceptor(
			server.AuthStreamInterceptor(jwtManager),
		),
	)
	pb.RegisterAuthServer(srv, authServer)
	pb.RegisterPasswordsServer(srv, passwordsServer)
	pb.RegisterCardsServer(srv, cardsServer)
	pb.RegisterNotesServer(srv, notesServer)

	fileDeletingWorkerPool.ProcessQueue()

	defer func() {
		fileDeletingWorkerPool.Close()

		log.Info().Msg("file deleting worker pool closing...")
		fileDeletingWorkerPool.Wait()
	}()

	log.Info().Msgf("server started on %s", cfg.Addr)

	if err := srv.Serve(listen); err != nil {
		log.Error().Err(err).Msg("server has been closed")
	}
}
