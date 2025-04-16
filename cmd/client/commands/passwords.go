package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

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
					return p.resolveConflict(ctx, conflictErr)
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
					return p.resolveConflict(ctx, conflictErr)
				}

				return err
			}

			return nil
		},
	}
}

func (p *PasswordsCommands) Delete() *cli.Command {
	return &cli.Command{
		Name: "delete",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			return p.passwordsUC.DeletePasswordByName(ctx, name)
		},
	}
}

func (p *PasswordsCommands) Sync() *cli.Command {
	return &cli.Command{
		Name: "sync",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			updates, err := p.passwordsUC.GetUpdates(ctx)
			if err != nil {
				return err
			}

			conflicts, operationErrors := p.applySyncUpdates(ctx, updates)

			for _, conflict := range conflicts {
				err := p.resolveConflict(ctx, conflict)
				if err != nil {
					operationErrors = append(operationErrors, err)
				}
			}

			w := bufio.NewWriter(cmd.Writer)

			if _, err = w.WriteString(fmt.Sprintf("Added: %d\n", len(updates.ToAdd))); err != nil {
				return nil
			}

			if _, err = w.WriteString(fmt.Sprintf("Updated: %d\n", len(updates.ToUpdate))); err != nil {
				return nil
			}

			if _, err = w.WriteString(fmt.Sprintf("Synced: %d\n", len(updates.ToSync))); err != nil {
				return nil
			}

			if len(operationErrors) > 0 {
				if _, err = w.WriteString("-------------------------\n"); err != nil {
					return err
				}

				for _, operationError := range operationErrors {
					if _, err = w.WriteString(fmt.Sprintf("Sync error: %s", operationError.Error())); err != nil {
						return err
					}
				}
			}

			return w.Flush()
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
	conflict *entity.PasswordConflictError,
) error {
	if conflict.Type() == entity.PasswordDeletedConflictType {
		return p.resolveDeleteConflict(ctx, conflict)
	}

	if conflict.Type() == entity.PasswordDiffConflictType {
		return p.resolveDiffConflict(ctx, conflict)
	}

	return ErrUnexpectedConflictType
}

func (p *PasswordsCommands) resolveDeleteConflict(
	ctx context.Context,
	conflict *entity.PasswordConflictError,
) error {
	shouldRecover, err := components.BoolPrompt(fmt.Sprintf(
		deleteConflictPromptTemplate,
		conflict.Actual().Name,
	))
	if err != nil {
		return err
	}

	if shouldRecover {
		conflict.Actual().Version = conflict.Incoming().Version + 1

		return p.passwordsUC.UpdatePassword(ctx, conflict.Actual())
	}

	return p.passwordsUC.DeletePasswordLocal(ctx, conflict.Actual().Name)
}

func (p *PasswordsCommands) resolveDiffConflict(
	ctx context.Context,
	conflict *entity.PasswordConflictError,
) error {
	key, err := p.keyProvider.Get(ctx)
	if err != nil {
		return err
	}

	if err = conflict.Actual().Open(key); err != nil {
		return err
	}

	if err = conflict.Incoming().Open(key); err != nil {
		return err
	}

	shouldOverride, err := components.BoolPrompt(fmt.Sprintf(
		diffConflictPromptTemplate,
		conflict.Incoming().Value, conflict.Incoming().Meta,
		conflict.Actual().Value, conflict.Actual().Meta,
	))
	if err != nil {
		return err
	}

	if shouldOverride {
		conflict.Actual().Version = conflict.Incoming().Version + 1

		return p.passwordsUC.UpdatePassword(ctx, conflict.Actual())
	}

	return p.passwordsUC.UpdatePasswordLocal(ctx, conflict.Incoming())
}

func (p *PasswordsCommands) applySyncUpdates(
	ctx context.Context,
	updates *usecase.PasswordsUpdates,
) ([]*entity.PasswordConflictError, []error) {
	var wg sync.WaitGroup

	type result struct {
		password *entity.Password
		err      error
	}

	operationsAmount := len(updates.ToUpdate) + len(updates.ToSync) + 1
	operationErrors := make([]error, 0, operationsAmount)
	conflicts := make([]*entity.PasswordConflictError, 0, len(updates.ToSync))
	resultChan := make(chan error, operationsAmount)

	if len(updates.ToAdd) > 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			resultChan <- p.passwordsUC.AddPasswordsLocal(ctx, updates.ToAdd)
		}()
	}

	if len(updates.ToUpdate) > 0 {
		wg.Add(len(updates.ToUpdate))

		for _, password := range updates.ToUpdate {
			go func(password *entity.Password) {
				defer wg.Done()

				resultChan <- p.passwordsUC.UpdatePasswordLocal(ctx, password)
			}(password)
		}
	}

	if len(updates.ToSync) > 0 {
		wg.Add(len(updates.ToSync))

		for _, password := range updates.ToSync {
			go func(password *entity.Password) {
				defer wg.Done()

				resultChan <- p.passwordsUC.UpdatePassword(ctx, password)
			}(password)
		}
	}

	go func() {
		wg.Wait()

		close(resultChan)
	}()

	for err := range resultChan {
		if err == nil {
			continue
		}

		var conflictErr *entity.PasswordConflictError

		if errors.As(err, &conflictErr) {
			conflicts = append(conflicts, conflictErr)
		} else {
			operationErrors = append(operationErrors, err)
		}
	}

	return conflicts, operationErrors
}
