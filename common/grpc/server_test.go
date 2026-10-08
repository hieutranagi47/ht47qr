package grpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"htqrcode/common"
	api "htqrcode/common/api/grpc"
	transport "htqrcode/common/grpc"
	qrcodeAPI "htqrcode/qrcode/api/grpc"
	"htqrcode/qrcode/app"
)

func TestErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err     error
		code    codes.Code
		message string
	}{
		{errors.New("database password=secret"), codes.Internal, "Internal Server Error"},
		{common.NewInvalidInputError("invalid_input", "bad request"), codes.InvalidArgument, "bad request"},
		{common.NewUnauthorizedError("unauthorized", "token required"), codes.Unauthenticated, "token required"},
		{common.NewForbiddenError("forbidden", "access denied"), codes.PermissionDenied, "access denied"},
		{common.NewNotFoundError("not_found", "missing"), codes.NotFound, "missing"},
		{common.NewConflictError("conflict", "conflict"), codes.Aborted, "conflict"},
		{context.Canceled, codes.Canceled, "Request canceled"},
		{context.DeadlineExceeded, codes.DeadlineExceeded, "Request deadline exceeded"},
	} {
		t.Run(tc.message, func(t *testing.T) {
			err := transport.Error(tc.err)
			require.Equal(t, tc.code, status.Code(err))
			require.Equal(t, tc.message, status.Convert(err).Message())
			require.NotContains(t, err.Error(), "secret")
		})
	}
	err := common.NewInvalidInputError("invalid_input", "bad field").WithDetails([]common.ErrorDetails{{EntityType: "request_field", EntityID: "message", ErrorSlug: "invalid_input", Message: "bad field"}}).WithInternalError(errors.New("secret"))
	detail := status.Convert(transport.Error(&err)).Details()[0].(*api.ErrorResponse)
	require.Equal(t, "message", detail.Details[0].EntityId)
	require.Equal(t, "invalid_input", detail.Slug)
	require.NoError(t, transport.Error(nil))
}

type rendererFunc func(context.Context, string, app.Options) ([]byte, error)

func (f rendererFunc) Render(ctx context.Context, payload string, options app.Options) ([]byte, error) {
	return f(ctx, payload, options)
}

func TestRecoveryAndBoundedShutdown(t *testing.T) {
	for _, panicHandler := range []bool{true, false} {
		t.Run(map[bool]string{true: "panic", false: "stuck handler"}[panicHandler], func(t *testing.T) {
			started := make(chan struct{})
			finished := make(chan struct{})
			renderer := rendererFunc(func(ctx context.Context, _ string, _ app.Options) ([]byte, error) {
				close(started)
				defer close(finished)
				if panicHandler {
					panic("private panic diagnostics")
				}
				<-ctx.Done()
				return nil, ctx.Err()
			})
			server := transport.NewServer()
			qrcodeAPI.Register(server, app.NewGenerator(renderer))
			listener := bufconn.Listen(1 << 20)
			served := make(chan error, 1)
			go func() { served <- server.Serve(listener) }()
			t.Cleanup(func() { server.Stop(); listener.Close(); require.NoError(t, <-served) })
			conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
			require.NoError(t, err)
			defer conn.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				_, err := qrcodeAPI.NewQRCodeServiceClient(conn).GenerateQRCode(ctx, &qrcodeAPI.GenerateQRCodeRequest{Message: &qrcodeAPI.MessageRequest{Text: "hello"}})
				result <- err
			}()
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("handler did not start")
			}
			if panicHandler {
				err := <-result
				require.Equal(t, codes.Internal, status.Code(err))
				require.NotContains(t, err.Error(), "private panic diagnostics")
			} else {
				stopped := make(chan struct{})
				go func() { server.Shutdown(20 * time.Millisecond); close(stopped) }()
				select {
				case <-stopped:
				case <-ctx.Done():
					t.Fatal("shutdown exceeded its bound")
				}
				require.Error(t, <-result)
			}
			select {
			case <-finished:
			case <-ctx.Done():
				t.Fatal("handler did not exit")
			}
		})
	}
}
