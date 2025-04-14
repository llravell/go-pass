package commands

import (
	"context"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

type NotesUseCase interface {
	UploadFile(ctx context.Context, name, meta string, file *os.File) error
	DownloadFile(ctx context.Context, name string, file *os.File) error
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

			err = commands.notesUC.UploadFile(ctx, name, meta, file)
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

			err = commands.notesUC.DownloadFile(ctx, name, file)
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
