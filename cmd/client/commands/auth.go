package commands

import (
	"context"
	"strings"

	cliController "github.com/llravell/go-pass/internal/controller/cli"
	"github.com/urfave/cli/v3"
)

func RegisterCommand(
	authController *cliController.AuthController,
) *cli.Command {
	return &cli.Command{
		Name: "register",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "login",
				Aliases:  []string{"l"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			login := strings.TrimSpace(c.String("login"))
			password := strings.TrimSpace(c.String("password"))

			err := authController.Register(ctx, login, password)
			if err != nil {
				return cli.Exit("registration failed", 1)
			}

			return nil
		},
	}
}

func LoginCommand(
	authController *cliController.AuthController,
) *cli.Command {
	return &cli.Command{
		Name: "login",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "login",
				Aliases:  []string{"l"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "password",
				Aliases:  []string{"p"},
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			login := strings.TrimSpace(c.String("login"))
			password := strings.TrimSpace(c.String("password"))

			err := authController.Register(ctx, login, password)
			if err != nil {
				return cli.Exit("login failed", 1)
			}

			return nil
		},
	}
}
