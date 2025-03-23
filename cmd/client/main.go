package main

import (
	"database/sql"
	"embed"
	"log"
	"os"
	"path"

	"github.com/pressly/goose/v3"
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

func main() {
	db, err := initStorage()
	if err != nil {
		log.Fatal(err)
	}

	err = runMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("init success")
}
