package tests

import (
	"bytes"
	"context"
	"image/png"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"htqrcode"
	errorsAPI "htqrcode/common/api/grpc"
	transport "htqrcode/common/grpc"
	api "htqrcode/qrcode/api/grpc"
)

func grpcClient(t *testing.T) api.QRCodeServiceClient {
	t.Helper()
	service, err := htqrcode.New(context.Background(), htqrcode.ExternalServices{})
	require.NoError(t, err)
	listener := bufconn.Listen(1 << 20)
	done := make(chan error, 1)
	go func() { done <- service.GRPCServer().Serve(listener) }()
	connection, err := grpc.NewClient("passthrough:///qrcode", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(transport.MaxResponseSize)))
	require.NoError(t, err)
	t.Cleanup(func() {
		connection.Close()
		service.GRPCServer().Stop()
		listener.Close()
		require.NoError(t, <-done)
	})
	return api.NewQRCodeServiceClient(connection)
}

func callContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestGRPCMessageTypesMatchREST(t *testing.T) {
	client := grpcClient(t)
	rest := startService(t)
	for _, tc := range qrMessageCases {
		t.Run(tc.name, func(t *testing.T) {
			message := &api.MessageRequest{}
			require.NoError(t, protojson.Unmarshal([]byte(tc.data), message))
			response, err := client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: message})
			require.NoError(t, err)
			require.Equal(t, "image/png", response.ContentType)
			restResponse := generate(t, rest, map[string]string{"data": tc.data}, nil)
			require.Equal(t, 200, restResponse.StatusCode)
			restImage, err := io.ReadAll(restResponse.Body)
			require.NoError(t, err)
			require.Equal(t, restImage, response.Image, "REST and gRPC must generate identical QR images")
		})
	}
}

func TestGRPCOptionsMatchREST(t *testing.T) {
	client, rest := grpcClient(t), startService(t)
	for _, shape := range []string{"square", "is_circle_shape", "is_custom_shape"} {
		t.Run(shape, func(t *testing.T) {
			options := &api.Options{QrWidth: proto.Uint32(12), BorderWidth: proto.Uint32(0), ForegroundColor: "#123456"}
			fields := map[string]string{"data": `{"text":"hello"}`, "qr_width": "12", "border_width": "0", "foreground_color": "#123456"}
			if shape != "square" {
				fields[shape] = "true"
			}
			options.IsCircleShape = shape == "is_circle_shape"
			options.IsCustomShape = shape == "is_custom_shape"
			response, err := client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: &api.MessageRequest{Text: "hello"}, Options: options})
			require.NoError(t, err)
			image, err := png.Decode(bytes.NewReader(response.Image))
			require.NoError(t, err)
			require.Equal(t, 252, image.Bounds().Dx())
			restImage, err := io.ReadAll(generate(t, rest, fields, nil).Body)
			require.NoError(t, err)
			require.Equal(t, restImage, response.Image)
		})
	}
	for _, jpeg := range []bool{false, true} {
		for _, logo := range []bool{false, true} {
			options := &api.Options{}
			field := "halftone_img"
			data := testImage(t, 20, jpeg)
			if logo {
				options.LogoImg = data
				field = "logo_img"
			} else {
				options.HalftoneImg = data
			}
			response, err := client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: &api.MessageRequest{Text: "hello"}, Options: options})
			require.NoError(t, err)
			restImage, err := io.ReadAll(generate(t, rest, map[string]string{"data": `{"text":"hello"}`}, map[string][]byte{field: data}).Body)
			require.NoError(t, err)
			require.Equal(t, restImage, response.Image)
		}
	}
}

func TestGRPCValidationAndSafeErrors(t *testing.T) {
	client := grpcClient(t)
	for _, tc := range []struct {
		name    string
		message *api.MessageRequest
		options *api.Options
		field   string
	}{
		{name: "missing message", field: "message"},
		{name: "missing text", message: &api.MessageRequest{}, field: "message"},
		{name: "missing wifi password", message: &api.MessageRequest{Type: "wifi", WifiName: "network"}, field: "message"},
		{name: "small width", options: &api.Options{QrWidth: proto.Uint32(5)}, field: "qr_width"},
		{name: "large width", options: &api.Options{QrWidth: proto.Uint32(256)}, field: "qr_width"},
		{name: "zero width", options: &api.Options{QrWidth: proto.Uint32(0)}, field: "qr_width"},
		{name: "large border", options: &api.Options{BorderWidth: proto.Uint32(1025)}, field: "border_width"},
		{name: "bad color", options: &api.Options{ForegroundColor: "oops"}, field: "foreground_color"},
		{name: "bad logo", options: &api.Options{LogoImg: []byte("invalid")}, field: "logo_img"},
		{name: "large logo", options: &api.Options{LogoImg: make([]byte, (1<<20)+1)}, field: "logo_img"},
		{name: "logo resolution", options: &api.Options{LogoImg: testImage(t, 501, false)}, field: "logo_img"},
		{name: "bad halftone", options: &api.Options{HalftoneImg: []byte("invalid")}, field: "halftone_img"},
		{name: "large halftone", options: &api.Options{HalftoneImg: make([]byte, (5<<20)+1)}, field: "halftone_img"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message := tc.message
			if tc.options != nil {
				message = &api.MessageRequest{Text: "hello"}
			}
			_, err := client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: message, Options: tc.options})
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			details := status.Convert(err).Details()
			require.Len(t, details, 1)
			detail := details[0].(*errorsAPI.ErrorResponse)
			require.Equal(t, "invalid_input", detail.Slug)
			require.EqualValues(t, 400, detail.HttpStatus)
			require.Len(t, detail.Details, 1)
			require.Equal(t, tc.field, detail.Details[0].EntityId)
		})
	}
	_, err := client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: &api.MessageRequest{Text: strings.Repeat("a", 4000)}})
	require.Equal(t, codes.Internal, status.Code(err))
	require.Equal(t, "failed to generate qr code", status.Convert(err).Message())
	detail := status.Convert(err).Details()[0].(*errorsAPI.ErrorResponse)
	require.Equal(t, "qr_code_generation_failed", detail.Slug)
	require.Empty(t, detail.Details)
	_, err = client.GenerateQRCode(callContext(t), &api.GenerateQRCodeRequest{Message: &api.MessageRequest{Text: strings.Repeat("a", transport.MaxRequestSize)}})
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestGRPCCorrelationAndCancellation(t *testing.T) {
	client := grpcClient(t)
	request := &api.GenerateQRCodeRequest{Message: &api.MessageRequest{Text: "hello"}}
	var header metadata.MD
	ctx := metadata.NewOutgoingContext(callContext(t), metadata.Pairs("correlation-id", "test-reference"))
	_, err := client.GenerateQRCode(ctx, request, grpc.Header(&header))
	require.NoError(t, err)
	require.Equal(t, []string{"test-reference"}, header.Get("correlation-id"))
	_, err = client.GenerateQRCode(callContext(t), request, grpc.Header(&header))
	require.NoError(t, err)
	require.NotEmpty(t, header.Get("correlation-id"))
	for _, refs := range [][]string{{"correlation-id", strings.Repeat("x", 129)}, {"correlation-id", "one", "correlation-id", "two"}} {
		ctx := metadata.NewOutgoingContext(callContext(t), metadata.Pairs(refs...))
		_, err := client.GenerateQRCode(ctx, request)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.GenerateQRCode(canceled, request)
	require.Equal(t, codes.Canceled, status.Code(err))
	expired, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	_, err = client.GenerateQRCode(expired, request)
	require.Equal(t, codes.DeadlineExceeded, status.Code(err))
}
