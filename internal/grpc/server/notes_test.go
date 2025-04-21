package server_test

import (
	"bytes"
	"context"
	"log"
	"net"
	"testing"

	"github.com/llravell/go-pass/internal/entity"
	"github.com/llravell/go-pass/internal/grpc/server"
	"github.com/llravell/go-pass/internal/mocks"
	usecase "github.com/llravell/go-pass/internal/usecase/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var (
	fileContent = []byte("test file content")
	fileSize    = int64(len(fileContent))
)

type fakeFileReader struct {
	*bytes.Reader
}

func (r *fakeFileReader) Close() error {
	return nil
}

type boomFileReader struct{}

func (r *boomFileReader) Read(_ []byte) (int, error) {
	return 0, errBoom
}

func (r *boomFileReader) Close() error {
	return errBoom
}

func startGRPCNotesServer(
	t *testing.T,
	filesRepo usecase.FilesRepository,
	filesS3Storage usecase.FilesS3Storage,
	fileDeletingWP usecase.FileDeletingWorkerPool,
) (pb.NotesClient, func()) {
	t.Helper()

	logger := zerolog.Nop()
	notesUsecase := usecase.NewFilesUseCase(filesRepo, filesS3Storage, fileDeletingWP, &logger)
	notesServer := server.NewNotesServer(notesUsecase, &logger)

	lis := bufconn.Listen(bufSize)
	server := grpc.NewServer(grpc.StreamInterceptor(fakeAuthStreamInterceptor))
	pb.RegisterNotesServer(server, notesServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	closeFn := func() {
		conn.Close()
		server.Stop()
	}

	return pb.NewNotesClient(conn), closeFn
}

func TestNotesServer_Upload(t *testing.T) {
	filesRepo := mocks.NewMockFilesRepository(gomock.NewController(t))
	s3Storage := mocks.NewMockFilesS3Storage(gomock.NewController(t))
	fileDeletingWP := mocks.NewMockFileDeletingWorkerPool(gomock.NewController(t))

	client, closeFn := startGRPCNotesServer(t, filesRepo, s3Storage, fileDeletingWP)

	defer closeFn()

	t.Run("upload file", func(t *testing.T) {
		filesRepo.EXPECT().
			UploadFile(gomock.Any(), defaultUserID, newFileMatcher("test", "notes"), gomock.Any(), gomock.Any()).
			Return(nil)

		stream, err := client.Upload(t.Context())
		require.NoError(t, err)

		err = stream.Send(&pb.FileChunk{
			Filename: "test",
			Data:     fileContent,
		})
		require.NoError(t, err)

		response, err := stream.CloseAndRecv()
		require.NoError(t, err)

		assert.True(t, response.GetSuccess())
	})

	t.Run("uploading file error", func(t *testing.T) {
		filesRepo.EXPECT().
			UploadFile(gomock.Any(), defaultUserID, newFileMatcher("test", "notes"), gomock.Any(), gomock.Any()).
			Return(errBoom)

		stream, err := client.Upload(t.Context())
		require.NoError(t, err)

		err = stream.Send(&pb.FileChunk{
			Filename: "test",
			Data:     fileContent,
		})
		require.NoError(t, err)

		_, err = stream.CloseAndRecv()
		assert.Error(t, err)
	})
}

//nolint:funlen
func TestNotesServer_Download(t *testing.T) {
	filesRepo := mocks.NewMockFilesRepository(gomock.NewController(t))
	s3Storage := mocks.NewMockFilesS3Storage(gomock.NewController(t))
	fileDeletingWP := mocks.NewMockFileDeletingWorkerPool(gomock.NewController(t))

	client, closeFn := startGRPCNotesServer(t, filesRepo, s3Storage, fileDeletingWP)

	defer closeFn()

	t.Run("download file", func(t *testing.T) {
		filesRepo.EXPECT().
			GetFileByName(gomock.Any(), defaultUserID, "notes", "test").
			Return(&entity.File{Name: "test", MinioBucket: "notes", Size: fileSize}, nil)

		s3Storage.EXPECT().
			DownloadFile(gomock.Any(), newFileMatcher("test", "notes")).
			Return(&fakeFileReader{bytes.NewReader(fileContent)}, nil)

		stream, err := client.Download(t.Context(), &pb.NotesDownloadRequest{Name: "test"})
		require.NoError(t, err)

		chunk, err := stream.Recv()
		require.NoError(t, err)

		assert.Equal(t, fileContent, chunk.GetData())
	})

	t.Run("file searching error", func(t *testing.T) {
		filesRepo.EXPECT().
			GetFileByName(gomock.Any(), defaultUserID, "notes", "test").
			Return(nil, errBoom)

		stream, err := client.Download(t.Context(), &pb.NotesDownloadRequest{Name: "test"})
		require.NoError(t, err)

		_, err = stream.Recv()
		assert.Error(t, err)
	})

	t.Run("file downloading error", func(t *testing.T) {
		filesRepo.EXPECT().
			GetFileByName(gomock.Any(), defaultUserID, "notes", "test").
			Return(&entity.File{Name: "test", MinioBucket: "notes", Size: fileSize}, nil)

		s3Storage.EXPECT().
			DownloadFile(gomock.Any(), newFileMatcher("test", "notes")).
			Return(nil, errBoom)

		stream, err := client.Download(t.Context(), &pb.NotesDownloadRequest{Name: "test"})
		require.NoError(t, err)

		_, err = stream.Recv()
		assert.Error(t, err)
	})

	t.Run("file reading error", func(t *testing.T) {
		filesRepo.EXPECT().
			GetFileByName(gomock.Any(), defaultUserID, "notes", "test").
			Return(&entity.File{Name: "test", MinioBucket: "notes", Size: fileSize}, nil)

		s3Storage.EXPECT().
			DownloadFile(gomock.Any(), newFileMatcher("test", "notes")).
			Return(&boomFileReader{}, nil)

		stream, err := client.Download(t.Context(), &pb.NotesDownloadRequest{Name: "test"})
		require.NoError(t, err)

		_, err = stream.Recv()
		assert.Error(t, err)
	})
}
