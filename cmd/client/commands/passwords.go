package commands

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/llravell/go-pass/cmd/client/components"
	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/client"
	"github.com/llravell/go-pass/pkg/encryption"
	"github.com/urfave/cli/v3"
)

type PasswordsCommands struct {
	passwordsUC *usecase.PasswordsUseCase
	keyProvider *components.EncryptionKeyProvider
}

func NewPasswordsCommands(
	passwordsUC *usecase.PasswordsUseCase,
	keyProvider *components.EncryptionKeyProvider,
) *PasswordsCommands {
	return &PasswordsCommands{
		passwordsUC: passwordsUC,
		keyProvider: keyProvider,
	}
}

func (p *PasswordsCommands) List() *cli.Command {
	return &cli.Command{
		Name: "list",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			passwords, err := p.passwordsUC.GetList(ctx)
			if err != nil {
				return err
			}

			if len(passwords) == 0 {
				cmd.Writer.Write([]byte("you don't have any passwords yet\n"))

				return nil
			}

			writer := bufio.NewWriter(cmd.Writer)

			for _, password := range passwords {
				writer.WriteString(password.Name + "\n")
			}

			writer.Flush()

			return nil
		},
	}
}

func (p *PasswordsCommands) Show() *cli.Command {
	return &cli.Command{
		Name: "show",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			pass, err := p.passwordsUC.GetPasswordByName(ctx, name)
			if err != nil {
				return err
			}

			key, err := p.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			rawPassword, err := key.Decrypt(pass.Value)
			if err != nil {
				return err
			}

			cmd.Writer.Write([]byte(rawPassword + "\n"))

			return nil
		},
	}
}

func (p *PasswordsCommands) Add() *cli.Command {
	return &cli.Command{
		Name: "add",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "meta",
				Aliases: []string{"m"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			password := strings.TrimSpace(cmd.Args().Get(1))
			meta := strings.TrimSpace(cmd.String("meta"))

			if len(name) == 0 || len(password) == 0 {
				return cli.Exit("got invalid args", 1)
			}

			key, err := p.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			return p.passwordsUC.AddNewPassword(ctx, key, name, password, meta)
		},
	}
}

func (p *PasswordsCommands) Edit() *cli.Command {
	return &cli.Command{
		Name: "edit",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			pass, err := p.passwordsUC.GetPasswordByName(ctx, name)
			if err != nil {
				return err
			}

			key, err := p.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			editText, err := p.buildPasswordEditText(pass, key)
			if err != nil {
				return err
			}

			updatedText, err := components.EditViaVI(editText)
			if err != nil {
				return err
			}

			err = p.parsePasswordEditText(updatedText, pass, key)
			if err != nil {
				return err
			}

			pass.BumpVersion()

			return p.passwordsUC.UpdatePassword(ctx, pass)
		},
	}
}

func (p *PasswordsCommands) buildPasswordEditText(
	password *entity.Password,
	key *encryption.Key,
) (string, error) {
	decryptedPass, err := key.Decrypt(password.Value)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n%s", decryptedPass, password.Meta), nil
}

func (p *PasswordsCommands) parsePasswordEditText(
	text string,
	password *entity.Password,
	key *encryption.Key,
) error {
	var rawPassword, meta string

	parts := strings.Split(text, "\n")
	if len(parts) == 0 {
		return cli.Exit("empty password", 1)
	}

	rawPassword = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		meta = strings.TrimSpace(parts[1])
	}

	encryptedPassword, err := key.Encrypt(rawPassword)
	if err != nil {
		return err
	}

	password.Value = encryptedPassword
	password.Meta = meta

	return nil
}
