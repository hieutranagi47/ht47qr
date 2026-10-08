# QR code REST and gRPC API

Both transports call the same `app.Generator` and renderer. All existing QR message types, colors, shapes, logos, and halftone options are supported.

- REST: `POST /api/v1/generate-qrcode`, returning PNG bytes.
- gRPC: `htqrcode.qrcode.v1.QRCodeService/GenerateQRCode`, returning `image` bytes and `content_type: image/png`.
- Contract and generated Go client: `qrcode/api/grpc/service.proto` and `qrcode/api/grpc` (`grpcapi`).

## Configuration

Set `SERVER_GRPC_PORT=8447` to start the gRPC listener alongside REST and SSE. Leave it empty to disable the gRPC listener. It uses the PEM contents in `TLS_CERT` and `TLS_KEY`, like the HTTPS listeners. gRPC requires TLS and HTTP/2. Existing callers of `Service.Run` can omit its optional final gRPC port argument.

The QR API has no authentication layer; REST and gRPC have the same access policy. Clients can send an optional `correlation-id` metadata value (at most 128 characters); the server returns a supplied or generated ID in response headers.

## REST example

```sh
curl --fail http://localhost:8888/api/v1/generate-qrcode \
  -F 'data={"type":"text","text":"https://example.com"}' \
  -F qr_width=10 \
  -F border_width=0 \
  -o qrcode.png
```

## gRPC example

Run from the repository root with the local development certificate:

```sh
grpcurl \
  -cacert external/cert/localhost+1.pem \
  -authority localhost \
  -import-path . \
  -proto qrcode/api/grpc/service.proto \
  -proto common/api/grpc/errors.proto \
  -H 'correlation-id: qr-example' \
  -d '{"message":{"type":"text","text":"https://example.com"},"options":{"qr_width":10,"border_width":0}}' \
  localhost:8447 htqrcode.qrcode.v1.QRCodeService/GenerateQRCode
```

`grpcurl` prints protobuf JSON: the `image` field contains base64-encoded PNG bytes. Pass the proto files explicitly; server reflection is not enabled.

Go clients can use the generated client:

```go
roots := x509.NewCertPool()
if !roots.AppendCertsFromPEM(caPEM) {
    return errors.New("invalid CA certificate")
}
conn, err := grpc.NewClient("localhost:8447",
    grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
        RootCAs: roots,
        ServerName: "localhost",
    })),
    grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(80<<20)),
)
if err != nil {
    return err
}
defer conn.Close()

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
response, err := grpcapi.NewQRCodeServiceClient(conn).GenerateQRCode(ctx,
    &grpcapi.GenerateQRCodeRequest{
        Message: &grpcapi.MessageRequest{Type: "text", Text: "https://example.com"},
        Options: &grpcapi.Options{BorderWidth: proto.Uint32(0)},
    },
)
if err != nil {
    return err
}
return os.WriteFile("qrcode.png", response.Image, 0644)
```

This snippet uses standard-library `context`, `crypto/tls`, `crypto/x509`, `errors`, `os`, and `time`; `google.golang.org/grpc`; `google.golang.org/grpc/credentials`; `google.golang.org/protobuf/proto`; and `grpcapi "htqrcode/qrcode/api/grpc"`. Supply the trusted certificate authority as `caPEM`. Increase the client's default 4 MiB receive limit for large output images, as shown.

## Validation and errors

Omitted `qr_width` and `border_width` default to 10. Explicit `border_width: 0` removes the border; explicit `qr_width: 0` is invalid. Width is limited to 6–255 and border to 0–1024. Logo and halftone fields carry PNG/JPEG bytes, with the same file size and resolution limits as REST. Circle mode takes precedence over custom shape and halftone rendering.

Serialized gRPC requests are limited to 8 MiB. Output dimensions are limited to 4096×4096 pixels. Invalid fields return `InvalidArgument`; oversized wire messages return `ResourceExhausted`; rendering failures return `Internal` with a safe public message. Cancellations and deadlines return the corresponding gRPC status codes. Internal diagnostics are logged, never returned to clients.

Application errors include a typed `htqrcode.common.v1.ErrorResponse` in `google.rpc.Status.details`, carrying the public message, slug, HTTP-equivalent status, and field details. Clients can inspect it with `status.Convert(err).Details()` and a type assertion to `*grpcapi.ErrorResponse` from `htqrcode/common/api/grpc`. gRPC's own transport errors, such as wire-size failures, use standard status codes.

## Regenerating protobuf code

Install `protoc` and the pinned Go plugins:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
```

Ensure `protoc` and the plugins (normally in `$(go env GOPATH)/bin`) are on `PATH`, then run:

```sh
go generate ./common/api/grpc ./qrcode/api/grpc
go test ./...
```

Tests compare REST and gRPC images for every supported message type, rendering options, PNG/JPEG inputs, and zero borders. They also cover validation, safe errors, request limits, correlation IDs, cancellation, TLS connections, listener failures, panic recovery, and bounded shutdown.
