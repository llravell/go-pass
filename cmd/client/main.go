package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"os"
	"path"

	"github.com/llravell/go-pass/cmd/client/commands"
	cliController "github.com/llravell/go-pass/internal/controller/cli"
	"github.com/llravell/go-pass/internal/repository"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"
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
	sessionRepo := repository.NewSessionSqliteRepository(db)
	authCliController := cliController.NewAuthController(sessionRepo)

	return &cli.Command{
		Name: "GOPASS",
		Commands: []*cli.Command{
			commands.RegisterCommand(authCliController),
			commands.LoginCommand(authCliController),
		},
	}
}

func main() {
	db, err := initStorage()
	if err != nil {
		log.Fatal(err)
	}

	err = runMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	cmd := buildCmd(db)

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
