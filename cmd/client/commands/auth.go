package commands

import (
	"context"
	"strings"

	"github.com/urfave/cli/v3"
)

type AuthUseCase interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) error
}

type AuthCommands struct {
	authUC AuthUseCase
}

func NewAuthCommands(authUC AuthUseCase) *AuthCommands {
	return &AuthCommands{
		authUC: authUC,
	}
}

func (commands *AuthCommands) Register() *cli.Command {
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

			err := commands.authUC.Register(ctx, login, password)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}
}

func (commands *AuthCommands) Login() *cli.Command {
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

			err := commands.authUC.Login(ctx, login, password)
			if err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}
}
