package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env"
)

const (
	_defaultAddr        = ":3200"
	_defaultDatabaseURI = ""
	_defaultJWTSecret   = "secret"
)

var ErrEmptyDatabaseURI = errors.New("got empty database uri")

type ServerConfig struct {
	Addr        string `env:"GRPC_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	JWTSecret   string `env:"JWT_SECRET"`
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Addr:        _defaultAddr,
		DatabaseURI: _defaultDatabaseURI,
		JWTSecret:   _defaultJWTSecret,
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server grpc address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "Database connect uri")
	flag.Parse()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *ServerConfig) Validate() error {
	if c.DatabaseURI == "" {
		return ErrEmptyDatabaseURI
	}

	return nil
}
