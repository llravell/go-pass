package client

import (
	"context"
	"errors"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (uc *NotesUseCase) UploadNote(
	ctx context.Context,
	name, meta string,
	reader io.Reader,
) error {
	stream, err := uc.notesClient.Upload(ctx)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024) // 1 KB
	isFirstChunk := true

	for {
		n, err := reader.Read(buffer)
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

func (uc *NotesUseCase) DownloadNote(
	ctx context.Context,
	name string,
	writer io.Writer,
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

		_, err = writer.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *NotesUseCase) GetNotes(
	ctx context.Context,
) ([]*entity.File, error) {
	response, err := uc.notesClient.GetList(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	notes := make([]*entity.File, 0, len(response.GetNotes()))

	for _, note := range response.GetNotes() {
		notes = append(notes, &entity.File{
			Name: note.GetName(),
			Meta: note.GetMeta(),
			Size: note.GetSize(),
		})
	}

	return notes, nil
}

func (uc *NotesUseCase) DeleteNote(
	ctx context.Context,
	name string,
) error {
	_, err := uc.notesClient.Delete(ctx, &pb.NotesDeleteRequest{Name: name})
	if err != nil {
		return err
	}

	return nil
}
