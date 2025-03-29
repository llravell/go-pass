package commands

import (
	"context"
	"strings"

	usecase "github.com/llravell/go-pass/internal/usecase/client"
	"github.com/urfave/cli/v3"
)

type AuthCommands struct {
	authUC *usecase.AuthUseCase
}

func NewAuthCommands(authUC *usecase.AuthUseCase) *AuthCommands {
	return &AuthCommands{
		authUC: authUC,
	}
}

func (auth *AuthCommands) Register() *cli.Command {
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

			err := auth.authUC.Register(ctx, login, password)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}
}

func (auth *AuthCommands) Login() *cli.Command {
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

			err := auth.authUC.Login(ctx, login, password)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}
}
