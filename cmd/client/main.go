package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"os"
	"path"

	"github.com/llravell/go-pass/cmd/client/commands"
	"github.com/llravell/go-pass/cmd/client/components"
	"github.com/llravell/go-pass/internal/repository"
	usecase "github.com/llravell/go-pass/internal/usecase/client"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const passDir = ".go_pass"
const dbName = "pass.db"

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

func initStorage() (*sql.DB, error) {
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	passDirPath := path.Join(homeDirPath, passDir)

	err = os.MkdirAll(passDirPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path.Join(passDirPath, dbName))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func buildCmd(db *sql.DB) *cli.Command {
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	sessionRepo := repository.NewSessionSqliteRepository(db)
	passwordsRepo := repository.NewPasswordsSqliteRepository(db)
	authClient := pb.NewAuthClient(conn)

	authUseCase := usecase.NewAuthUseCase(sessionRepo, authClient)
	passwordsUseCase := usecase.NewPasswordsUseCase(passwordsRepo)

	masterPassPrompter := components.NewMasterPassPrompter(authUseCase)
	authCommands := commands.NewAuthCommands(authUseCase)
	passwordsCommands := commands.NewPasswordsCommands(passwordsUseCase, masterPassPrompter)

	return &cli.Command{
		Name: "GOPASS",
		Commands: []*cli.Command{
			authCommands.Login(),
			authCommands.Register(),

			&cli.Command{
				Name: "init",
				Action: func(context.Context, *cli.Command) error {
					return runMigrations(db)
				},
			},
			&cli.Command{
				Name: "passwords",
				Commands: []*cli.Command{
					passwordsCommands.Add(),
					passwordsCommands.Edit(),
				},
			},
		},
		After: func(context.Context, *cli.Command) error {
			return conn.Close()
		},
	}
}

func main() {
	db, err := initStorage()
	if err != nil {
		log.Fatal(err)
	}

	cmd := buildCmd(db)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
