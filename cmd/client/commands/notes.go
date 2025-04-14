package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/urfave/cli/v3"
)

type NotesUseCase interface {
	UploadNote(ctx context.Context, name, meta string, file *os.File) error
	DownloadNote(ctx context.Context, name string, file *os.File) error
	GetNotes(ctx context.Context) ([]*entity.File, error)
}

type NotesCommands struct {
	notesUC NotesUseCase
}

func NewNotesCommands(notesUC NotesUseCase) *NotesCommands {
	return &NotesCommands{
		notesUC: notesUC,
	}
}

func (commands *NotesCommands) Upload() *cli.Command {
	return &cli.Command{
		Name: "upload",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "meta",
				Aliases: []string{"m"},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			filepath := strings.TrimSpace(cmd.Args().Get(1))
			meta := strings.TrimSpace(cmd.String("meta"))

			if len(name) == 0 || len(filepath) == 0 {
				return cli.Exit("got invalid args", 1)
			}

			file, err := os.Open(filepath)
			if err != nil {
				return err
			}

			defer file.Close()

			if _, err = cmd.Writer.Write([]byte("start uploading...\n")); err != nil {
				return err
			}

			err = commands.notesUC.UploadNote(ctx, name, meta, file)
			if err != nil {
				return err
			}

			if _, err = cmd.Writer.Write([]byte("done!\n")); err != nil {
				return err
			}

			return nil
		},
	}
}

func (commands *NotesCommands) Download() *cli.Command {
	return &cli.Command{
		Name: "download",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := strings.TrimSpace(cmd.Args().Get(0))
			dst := strings.TrimSpace(cmd.Args().Get(1))

			if len(name) == 0 || len(dst) == 0 {
				return cli.Exit("got invalid args", 1)
			}

			file, err := os.Create(dst)
			if err != nil {
				return err
			}

			defer file.Close()

			if _, err = cmd.Writer.Write([]byte("start downloading...\n")); err != nil {
				return err
			}

			err = commands.notesUC.DownloadNote(ctx, name, file)
			if err != nil {
				return err
			}

			if _, err = cmd.Writer.Write([]byte("done!\n")); err != nil {
				return err
			}

			return nil
		},
	}
}

func (commands *NotesCommands) List() *cli.Command {
	return &cli.Command{
		Name: "list",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			notes, err := commands.notesUC.GetNotes(ctx)
			if err != nil {
				return err
			}

			if len(notes) == 0 {
				_, err := cmd.Writer.Write([]byte("you dont have any notes\n"))

				return err
			}

			writter := bufio.NewWriter(cmd.Writer)

			for i, note := range notes {
				_, err = fmt.Fprintf(writter, "%d. %s (%db)\n", i+1, note.Name, note.Size)
				if err != nil {
					return err
				}
			}

			return writter.Flush()
		},
	}
}
