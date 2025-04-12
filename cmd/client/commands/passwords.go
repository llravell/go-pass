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

type PasswordsUseCase interface {
	GetList(ctx context.Context) ([]*entity.Password, error)
	GetPasswordByName(ctx context.Context, name string) (*entity.Password, error)
	UpdatePassword(ctx context.Context, password *entity.Password) error
	UpdatePasswordLocal(ctx context.Context, password *entity.Password) error
	DeletePasswordLocal(ctx context.Context, name string) error
	AddPasswordsLocal(ctx context.Context, passwords []*entity.Password) error
	GetUpdates(ctx context.Context) (*entity.PasswordsUpdates, error)
	AddNewPassword(ctx context.Context, password entity.Password) error
	DeletePasswordByName(ctx context.Context, name string) error
}

type PasswordsCommands struct {
	passwordsUC PasswordsUseCase
	keyProvider *components.EncryptionKeyProvider
}

func NewPasswordsCommands(
	passwordsUC PasswordsUseCase,
	keyProvider *components.EncryptionKeyProvider,
) *PasswordsCommands {
	return &PasswordsCommands{
		passwordsUC: passwordsUC,
		keyProvider: keyProvider,
	}
}

func (commands *PasswordsCommands) List() *cli.Command {
	return &cli.Command{
		Name: "list",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			passwords, err := commands.passwordsUC.GetList(ctx)
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

func (commands *PasswordsCommands) Show() *cli.Command {
	return &cli.Command{
		Name: "show",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			pass, err := commands.passwordsUC.GetPasswordByName(ctx, name)
			if err != nil {
				return err
			}

			key, err := commands.keyProvider.Get(ctx)
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

func (commands *PasswordsCommands) Add() *cli.Command {
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

			key, err := commands.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			if err = password.Close(key); err != nil {
				return err
			}

			err = commands.passwordsUC.AddNewPassword(ctx, password)
			if err != nil {
				var conflictErr *entity.PasswordConflictError

				if errors.As(err, &conflictErr) {
					return commands.resolveConflict(ctx, conflictErr)
				}

				return err
			}

			return nil
		},
	}
}

func (commands *PasswordsCommands) Edit() *cli.Command {
	return &cli.Command{
		Name: "edit",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			pass, err := commands.passwordsUC.GetPasswordByName(ctx, name)
			if err != nil {
				return err
			}

			key, err := commands.keyProvider.Get(ctx)
			if err != nil {
				return err
			}

			if err = pass.Open(key); err != nil {
				return err
			}

			updatedText, err := components.EditViaVI(commands.buildPasswordEditText(pass))
			if err != nil {
				return err
			}

			if err = commands.parsePasswordEditText(updatedText, pass); err != nil {
				return err
			}

			if err = pass.Close(key); err != nil {
				return err
			}

			pass.BumpVersion()

			err = commands.passwordsUC.UpdatePassword(ctx, pass)
			if err != nil {
				var conflictErr *entity.PasswordConflictError

				if errors.As(err, &conflictErr) {
					return commands.resolveConflict(ctx, conflictErr)
				}

				return err
			}

			return nil
		},
	}
}

func (commands *PasswordsCommands) Delete() *cli.Command {
	return &cli.Command{
		Name: "delete",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			if len(name) == 0 {
				return cli.Exit("got empty name", 1)
			}

			return commands.passwordsUC.DeletePasswordByName(ctx, name)
		},
	}
}

//nolint:cyclop
func (commands *PasswordsCommands) Sync() *cli.Command {
	return &cli.Command{
		Name: "sync",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			updates, err := commands.passwordsUC.GetUpdates(ctx)
			if err != nil {
				return err
			}

			conflicts, operationErrors := commands.applySyncUpdates(ctx, updates)

			for _, conflict := range conflicts {
				err := commands.resolveConflict(ctx, conflict)
				if err != nil {
					operationErrors = append(operationErrors, err)
				}
			}

			writter := bufio.NewWriter(cmd.Writer)

			if _, err = fmt.Fprintf(writter, "Added: %d\n", len(updates.ToAdd)); err != nil {
				return err
			}

			if _, err = fmt.Fprintf(writter, "Updated: %d\n", len(updates.ToUpdate)); err != nil {
				return err
			}

			if _, err = fmt.Fprintf(writter, "Synced: %d\n", len(updates.ToSync)); err != nil {
				return err
			}

			if len(operationErrors) == 0 {
				return writter.Flush()
			}

			if _, err = writter.WriteString("-------------------------\n"); err != nil {
				return err
			}

			for _, operationError := range operationErrors {
				if _, err = writter.WriteString("Sync error: " + operationError.Error()); err != nil {
					return err
				}
			}

			return writter.Flush()
		},
	}
}

func (commands *PasswordsCommands) buildPasswordEditText(
	password *entity.Password,
) string {
	return fmt.Sprintf("%s\n%s", password.Value, password.Meta)
}

func (commands *PasswordsCommands) parsePasswordEditText(
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

func (commands *PasswordsCommands) resolveConflict(
	ctx context.Context,
	conflict *entity.PasswordConflictError,
) error {
	if conflict.Type() == entity.PasswordDeletedConflictType {
		return commands.resolveDeleteConflict(ctx, conflict)
	}

	if conflict.Type() == entity.PasswordDiffConflictType {
		return commands.resolveDiffConflict(ctx, conflict)
	}

	return ErrUnexpectedConflictType
}

func (commands *PasswordsCommands) resolveDeleteConflict(
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

		return commands.passwordsUC.UpdatePassword(ctx, conflict.Actual())
	}

	return commands.passwordsUC.DeletePasswordLocal(ctx, conflict.Actual().Name)
}

func (commands *PasswordsCommands) resolveDiffConflict(
	ctx context.Context,
	conflict *entity.PasswordConflictError,
) error {
	key, err := commands.keyProvider.Get(ctx)
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

		return commands.passwordsUC.UpdatePassword(ctx, conflict.Actual())
	}

	return commands.passwordsUC.UpdatePasswordLocal(ctx, conflict.Incoming())
}

//nolint:funlen
func (commands *PasswordsCommands) applySyncUpdates(
	ctx context.Context,
	updates *entity.PasswordsUpdates,
) ([]*entity.PasswordConflictError, []error) {
	var wg sync.WaitGroup

	operationsAmount := len(updates.ToUpdate) + len(updates.ToSync) + 1
	operationErrors := make([]error, 0, operationsAmount)
	conflicts := make([]*entity.PasswordConflictError, 0, len(updates.ToSync))
	resultChan := make(chan error, operationsAmount)

	if len(updates.ToAdd) > 0 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			resultChan <- commands.passwordsUC.AddPasswordsLocal(ctx, updates.ToAdd)
		}()
	}

	if len(updates.ToUpdate) > 0 {
		wg.Add(len(updates.ToUpdate))

		for _, password := range updates.ToUpdate {
			go func(password *entity.Password) {
				defer wg.Done()

				resultChan <- commands.passwordsUC.UpdatePasswordLocal(ctx, password)
			}(password)
		}
	}

	if len(updates.ToSync) > 0 {
		wg.Add(len(updates.ToSync))

		for _, password := range updates.ToSync {
			go func(password *entity.Password) {
				defer wg.Done()

				resultChan <- commands.passwordsUC.UpdatePassword(ctx, password)
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
