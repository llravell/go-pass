package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env"
)

const (
	_defaultAddr                 = ":3200"
	_defaultDatabaseURI          = ""
	_defaultJWTSecret            = "secret"
	_defaultMinioAddr            = "localhost:9000"
	_defaultMinioAccessKeyID     = ""
	_defaultMinioSecretAccessKey = ""
)

var ErrEmptyDatabaseURI = errors.New("got empty database uri")

type ServerConfig struct {
	Addr                 string `env:"GRPC_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	JWTSecret            string `env:"JWT_SECRET"`
	MinioAddr            string `env:"MINIO_ADDR"`
	MinioAccessKeyID     string `env:"MINIO_ACCESS_KEY_ID"`
	MinioSecretAccessKey string `env:"MINIO_SECRET_ACCESS_KEY"`
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{
		Addr:                 _defaultAddr,
		DatabaseURI:          _defaultDatabaseURI,
		JWTSecret:            _defaultJWTSecret,
		MinioAddr:            _defaultMinioAddr,
		MinioAccessKeyID:     _defaultMinioAccessKeyID,
		MinioSecretAccessKey: _defaultMinioSecretAccessKey,
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "Server grpc address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "Database connect uri")
	flag.StringVar(&cfg.MinioAddr, "minio-addr", cfg.MinioAddr, "Minio connect uri")
	flag.StringVar(&cfg.MinioAccessKeyID, "minio-access-key", cfg.MinioAccessKeyID, "Minio access key id")
	flag.StringVar(&cfg.MinioSecretAccessKey, "minio-secret", cfg.MinioSecretAccessKey, "Minio access key secret")
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
