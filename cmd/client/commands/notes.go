package commands

import (
	"context"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

type NotesUseCase interface {
	UploadFile(ctx context.Context, name, meta string, file *os.File) error
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

			if _, err = cmd.Writer.Write([]byte("start uploading...")); err != nil {
				return err
			}

			err = commands.notesUC.UploadFile(ctx, name, meta, file)
			if err != nil {
				return err
			}

			if _, err = cmd.Writer.Write([]byte("done!")); err != nil {
				return err
			}

			return nil
		},
	}
}
