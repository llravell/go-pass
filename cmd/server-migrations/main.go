package main

import (
	"database/sql"
	"embed"
	"flag"
	"log"

	"github.com/caarlos0/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*sql
var embedMigrations embed.FS

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}

type Opts struct {
	DatabaseURI string `env:"DATABASE_URI"`
}

func main() {
	var opts Opts

	if err := env.Parse(&opts); err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&opts.DatabaseURI, "d", opts.DatabaseURI, "Database connect uri")
	flag.Parse()

	if len(opts.DatabaseURI) == 0 {
		log.Fatal("got empty database uri")
	}

	db, err := sql.Open("pgx", opts.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = runMigrations(db)
	if err != nil {
		log.Println(err)
	}
}
