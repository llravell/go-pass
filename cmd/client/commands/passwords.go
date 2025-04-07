package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/llravell/go-pass/cmd/client/components"
	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/client"
	"github.com/urfave/cli/v3"
)

var ErrUnexpectedConflictType = errors.New("got unexpected conflict type")

var deleteConflictPromptTemplate = `
Password "%s" has been deleted.
Do you want to recover it?
`
var diffConflictPromptTemplate = `
Got conflict while sync.
----------------------
Server:
%s
%s
----------------------
Local:
%s
%s
----------------------
Do you want to override server version?
`

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
				_, err = cmd.Writer.Write([]byte("you don't have any passwords yet\n"))

				return err
			}

			writer := bufio.NewWriter(cmd.Writer)

			for _, password := range passwords {
				_, err = writer.WriteString(password.Name + "\n")
				if err != nil {
					return err
				}
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

			if err = pass.Open(key); err != nil {
				return err
			}

			_, err = cmd.Writer.Write([]byte(pass.Value + "\n"))

			return err
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
			rawPassword := strings.TrimSpace(cmd.Args().Get(1))
			meta := strings.TrimSpace(cmd.String("meta"))

			if len(name) == 0 || len(rawPassword) == 0 {
				return cli.Exit("got invalid args", 1)
			}

			password := entity.Password{
				Name:    name,
				Value:   rawPassword,
				Meta:    meta,
				Version: 1,
			}

			key, err := p.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			if err = password.Close(key); err != nil {
				return err
			}

			err = p.passwordsUC.AddNewPassword(ctx, password)
			if err != nil {
				var conflictErr *entity.PasswordConflictError

				if errors.As(err, &conflictErr) {
					return p.resolveConflict(ctx, &password, conflictErr)
				}

				return err
			}

			return nil
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

			if err = pass.Open(key); err != nil {
				return err
			}

			updatedText, err := components.EditViaVI(p.buildPasswordEditText(pass))
			if err != nil {
				return err
			}

			if err = p.parsePasswordEditText(updatedText, pass); err != nil {
				return err
			}

			if err = pass.Close(key); err != nil {
				return err
			}

			pass.BumpVersion()

			err = p.passwordsUC.UpdatePassword(ctx, pass)
			if err != nil {
				var conflictErr *entity.PasswordConflictError

				if errors.As(err, &conflictErr) {
					return p.resolveConflict(ctx, pass, conflictErr)
				}

				return err
			}

			return nil
		},
	}
}

func (p *PasswordsCommands) buildPasswordEditText(
	password *entity.Password,
) string {
	return fmt.Sprintf("%s\n%s", password.Value, password.Meta)
}

func (p *PasswordsCommands) parsePasswordEditText(
	text string,
	password *entity.Password,
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

	password.Value = rawPassword
	password.Meta = meta

	return nil
}

func (p *PasswordsCommands) resolveConflict(
	ctx context.Context,
	targetPassword *entity.Password,
	conflictErr *entity.PasswordConflictError,
) error {
	if conflictErr.Type() == entity.PasswordDeletedConflictType {
		return p.resolveDeleteConflict(ctx, targetPassword, conflictErr.Password())
	}

	if conflictErr.Type() == entity.PasswordDiffConflictType {
		return p.resolveDiffConflict(ctx, targetPassword, conflictErr.Password())
	}

	return ErrUnexpectedConflictType
}

func (p *PasswordsCommands) resolveDeleteConflict(
	ctx context.Context,
	password *entity.Password,
	conflictedPassword *entity.Password,
) error {
	shouldRecover, err := components.BoolPrompt(fmt.Sprintf(
		deleteConflictPromptTemplate,
		password.Name,
	))
	if err != nil {
		return err
	}

	if shouldRecover {
		password.Version = conflictedPassword.Version + 1

		return p.passwordsUC.UpdatePassword(ctx, password)
	}

	return p.passwordsUC.DeletePasswordLocal(ctx, password)
}

func (p *PasswordsCommands) resolveDiffConflict(
	ctx context.Context,
	password *entity.Password,
	conflictedPassword *entity.Password,
) error {
	key, err := p.keyProvider.Get(ctx)
	if err != nil {
		return err
	}

	if err = password.Open(key); err != nil {
		return err
	}

	if err = conflictedPassword.Open(key); err != nil {
		return err
	}

	shouldOverride, err := components.BoolPrompt(fmt.Sprintf(
		diffConflictPromptTemplate,
		conflictedPassword.Value, conflictedPassword.Meta,
		password.Value, password.Meta,
	))
	if err != nil {
		return err
	}

	if shouldOverride {
		password.Version = conflictedPassword.Version + 1

		return p.passwordsUC.UpdatePassword(ctx, password)
	}

	return p.passwordsUC.UpdatePasswordLocal(ctx, conflictedPassword)
}
