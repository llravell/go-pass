package client

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
)

type NotesUseCase struct {
	notesClient pb.NotesClient
}

func NewNotesUseCase(
	notesClient pb.NotesClient,
) *NotesUseCase {
	return &NotesUseCase{
		notesClient: notesClient,
	}
}

func (uc *NotesUseCase) UploadFile(
	ctx context.Context,
	name, meta string,
	file *os.File,
) error {
	stream, err := uc.notesClient.Upload(ctx)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024) // 1 KB
	isFirstChunk := true

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		chunk := &pb.FileChunk{
			Data: buffer[:n],
		}

		if isFirstChunk {
			chunk.Filename = name
			chunk.Meta = meta
			isFirstChunk = false
		}

		if err := stream.Send(chunk); err != nil {
			return err
		}
	}

	result, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	if !result.GetSuccess() {
		return entity.ErrFileUploadingFailed
	}

	return nil
}

func (uc *NotesUseCase) DownloadFile(
	ctx context.Context,
	name string,
	file *os.File,
) error {
	stream, err := uc.notesClient.Download(ctx, &pb.NotesDownloadRequest{Name: name})
	if err != nil {
		return err
	}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		data := chunk.GetData()

		_, err = file.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}
