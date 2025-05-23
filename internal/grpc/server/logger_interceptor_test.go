package server_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/llravell/go-pass/internal/grpc/server"
	pb "github.com/llravell/go-pass/pkg/grpc"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	out := &bytes.Buffer{}
	logger := zerolog.New(out)

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	client, closeFn := startGRPCEchoServer(
		t,
		logging.UnaryServerInterceptor(server.Logger(&logger), opts...),
	)
	defer closeFn()

	t.Run("interceptor writes logs", func(t *testing.T) {
		_, err := client.Send(t.Context(), &pb.Message{})
		require.NoError(t, err)

		var startedCallLog, finishedCallLog map[string]string

		logs := bytes.SplitN(out.Bytes(), []byte{'\n'}, 2)
		require.Len(t, logs, 2)

		err = json.NewDecoder(bytes.NewReader(logs[0])).Decode(&startedCallLog)
		require.NoError(t, err)
		err = json.NewDecoder(bytes.NewReader(logs[1])).Decode(&finishedCallLog)
		require.NoError(t, err)

		assert.Equal(t, "info", startedCallLog["level"])
		assert.EqualValues(t, startedCallLog["level"], finishedCallLog["level"])

		assert.Equal(t, "echo.Echo", startedCallLog["grpc.service"])
		assert.EqualValues(t, startedCallLog["grpc.service"], finishedCallLog["grpc.service"])

		assert.Equal(t, "Send", startedCallLog["grpc.method"])
		assert.EqualValues(t, startedCallLog["grpc.method"], finishedCallLog["grpc.method"])

		assert.Equal(t, "OK", finishedCallLog["grpc.code"])
	})
}
