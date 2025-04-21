package server

import (
	"context"
	"errors"
	"io"

	"github.com/llravell/go-pass/internal/entity"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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

type FilesUseCase interface {
	UploadFile(
		ctx context.Context,
		userID int,
		file *entity.File,
		fileReader io.Reader,
	) error
	DownloadFile(
		ctx context.Context,
		userID int,
		bucket string,
		name string,
	) (io.ReadCloser, error)
	GetFiles(
		ctx context.Context,
		userID int,
		bucket string,
	) ([]*entity.File, error)
	DeleteFile(
		ctx context.Context,
		userID int,
		bucket string,
		name string,
	) error
}

type NotesServer struct {
	pb.UnimplementedNotesServer

	filesUC FilesUseCase
	log     *zerolog.Logger
}

func NewNotesServer(
	filesUC FilesUseCase,
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

func (s *NotesServer) GetList(ctx context.Context, _ *emptypb.Empty) (*pb.NotesListResponse, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	notes, err := s.filesUC.GetFiles(ctx, userID, notesMinioBucket)
	if err != nil {
		s.log.Error().Err(err).Msg("fetching notes failed")

		return nil, status.Error(codes.Internal, "fetching notes failed")
	}

	response := &pb.NotesListResponse{
		Notes: make([]*pb.FileInfo, 0, len(notes)),
	}

	for _, note := range notes {
		response.Notes = append(response.Notes, &pb.FileInfo{
			Name: note.Name,
			Meta: note.Meta,
			Size: note.Size,
		})
	}

	return response, nil
}

func (s *NotesServer) Delete(ctx context.Context, in *pb.NotesDeleteRequest) (*emptypb.Empty, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok {
		s.log.Error().Msg("getting userID from ctx failed")

		return nil, status.Error(codes.Unauthenticated, "failed to resolve user id")
	}

	err := s.filesUC.DeleteFile(ctx, userID, notesMinioBucket, in.GetName())
	if err != nil {
		s.log.Error().Err(err).Msg("deleting note failed")

		return nil, status.Error(codes.Internal, "deleting note failed")
	}

	return &emptypb.Empty{}, nil
}
