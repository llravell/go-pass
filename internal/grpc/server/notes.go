package server

import (
	"errors"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const notesMinioBucket = "notes"

type noteUploadReader struct {
	buf    []byte
	stream grpc.ClientStreamingServer[pb.FileChunk, pb.NotesUploadResponse]
}

func (r *noteUploadReader) Read(p []byte) (int, error) {
	if len(r.buf) == 0 {
		chunk, err := r.stream.Recv()
		if err != nil {
			return 0, err
		}

		r.buf = chunk.GetData()
	}

	n := copy(p, r.buf)
	r.buf = r.buf[n:]

	return n, nil
}

type NotesServer struct {
	pb.UnimplementedNotesServer

	filesUC *usecase.FilesUseCase
	log     *zerolog.Logger
}

func NewNotesServer(
	filesUC *usecase.FilesUseCase,
	log *zerolog.Logger,
) *NotesServer {
	return &NotesServer{
		filesUC: filesUC,
		log:     log,
	}
}

func (s *NotesServer) Upload(stream grpc.ClientStreamingServer[pb.FileChunk, pb.NotesUploadResponse]) error {
	userID, ok := GetUserIDFromContext(stream.Context())
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	firstChunk, err := stream.Recv()
	if err != nil {
		return err
	}

	file := &entity.File{
		Name:        firstChunk.GetFilename(),
		Meta:        firstChunk.GetMeta(),
		MinioBucket: notesMinioBucket,
	}

	reader := &noteUploadReader{
		buf:    firstChunk.GetData(),
		stream: stream,
	}

	err = s.filesUC.UploadFile(stream.Context(), userID, file, reader)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&pb.NotesUploadResponse{
		Success: true,
	})
}

func (s *NotesServer) Download(in *pb.NotesDownloadRequest, stream grpc.ServerStreamingServer[pb.FileChunk]) error {
	userID, ok := GetUserIDFromContext(stream.Context())
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	fileReader, err := s.filesUC.DownloadFile(stream.Context(), userID, notesMinioBucket, in.GetName())
	if err != nil {
		return err
	}

	defer fileReader.Close()

	buffer := make([]byte, 1024) // 1 KB

	for {
		n, err := fileReader.Read(buffer)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}

			if n == 0 {
				break
			}
		}

		chunk := &pb.FileChunk{
			Data: buffer[:n],
		}

		if err := stream.Send(chunk); err != nil {
			return err
		}
	}

	return nil
}
