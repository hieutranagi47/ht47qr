// Package grpc provides shared transport limits, correlation IDs and safe errors.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"htqrcode/common"
	api "htqrcode/common/api/grpc"
	"htqrcode/common/log"
)

const MaxRequestSize = 8 << 20

// A 4096x4096 PNG can exceed gRPC's default 4 MiB response limit.
const MaxResponseSize = 80 << 20

type Server struct{ *grpc.Server }

func NewServer(options ...grpc.ServerOption) *Server {
	options = append(options, grpc.MaxRecvMsgSize(MaxRequestSize), grpc.MaxSendMsgSize(MaxResponseSize), grpc.MaxConcurrentStreams(128), grpc.ChainUnaryInterceptor(interceptor))
	return &Server{Server: grpc.NewServer(options...)}
}

func interceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (out any, err error) {
	defer func() {
		if recover() != nil {
			log.FromContext(ctx).Error("gRPC handler panicked", "method", info.FullMethod)
			out, err = nil, Error(fmt.Errorf("handler panic"))
		}
	}()
	md, _ := metadata.FromIncomingContext(ctx)
	refs := md.Get("correlation-id")
	if len(refs) > 1 || (len(refs) == 1 && len(refs[0]) > 128) {
		return nil, Error(common.NewInvalidInputError("invalid_input", "Invalid correlation reference"))
	}
	ref := uuid.NewString()
	if len(refs) == 1 && refs[0] != "" {
		ref = refs[0]
	}
	ctx = log.ContextWithCorrelationID(ctx, ref)
	_ = grpc.SetHeader(ctx, metadata.Pairs("correlation-id", ref))
	out, err = next(ctx, req)
	if err != nil {
		log.FromContext(ctx).Error("gRPC request failed", "method", info.FullMethod, "error", err)
		return nil, Error(err)
	}
	return out, nil
}

// Error attaches the public contract as a typed status detail, never err.Error().
func Error(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "Request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "Request deadline exceeded")
	}
	public, httpStatus := common.PublicHTTPError(err)
	code := codes.Internal
	switch httpStatus {
	case 400, 422:
		code = codes.InvalidArgument
	case 401:
		code = codes.Unauthenticated
	case 403:
		code = codes.PermissionDenied
	case 404:
		code = codes.NotFound
	case 409:
		code = codes.Aborted
	case 410, 423:
		code = codes.FailedPrecondition
	case 413, 429:
		code = codes.ResourceExhausted
	case 502, 503:
		code = codes.Unavailable
	case 504:
		code = codes.DeadlineExceeded
	}
	detail := &api.ErrorResponse{Message: public.Message, Slug: public.Slug, HttpStatus: int32(httpStatus), Details: []*api.ErrorDetail{}}
	for _, d := range public.Details {
		detail.Details = append(detail.Details, &api.ErrorDetail{EntityType: d.EntityType, EntityId: d.EntityID, ErrorSlug: d.ErrorSlug, Message: d.Message})
	}
	st, e := status.New(code, public.Message).WithDetails(detail)
	if e != nil {
		return status.Error(codes.Internal, "Internal Server Error")
	}
	return st.Err()
}

// Shutdown bounds draining so stuck external calls cannot block process shutdown.
func (s *Server) Shutdown(timeout time.Duration) {
	done := make(chan struct{})
	go func() { s.GracefulStop(); close(done) }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		s.Stop()
		<-done
	}
}
