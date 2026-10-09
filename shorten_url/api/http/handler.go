package http

import (
	"context"
	"errors"
	"strings"

	"htqrcode/common"
	"htqrcode/shorten_url/app"
	"htqrcode/shorten_url/domain"
)

type Handler struct{ service *app.Service }

func NewHandler(service *app.Service) Handler {
	if service == nil {
		panic("short URL service is required")
	}
	return Handler{service: service}
}

func Register(_ context.Context, e common.EchoRouter, h Handler) error {
	RegisterHandlers(e, NewStrictHandler(h, nil))
	return nil
}

func (h Handler) CreateShortenURL(ctx context.Context, request CreateShortenURLRequestObject) (CreateShortenURLResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("request_body_required", "request body is required")
	}
	if strings.TrimSpace(request.Params.IdempotencyKey) == "" {
		return nil, common.NewInvalidInputError("idempotency_key_required", "Idempotency-Key is required")
	}
	result, err := h.service.Create(ctx, app.CreateInput{
		LongURL: request.Body.LongUrl, IdempotencyKey: request.Params.IdempotencyKey,
	})
	if err != nil {
		return nil, createErrorResponse(err)
	}
	return CreateShortenURL201JSONResponse{
		ShortUrl:  "/r/" + result.URL.Code(),
		LongUrl:   result.URL.LongURL(),
		ExpiresAt: result.URL.ExpiresAt(),
	}, nil
}

func (h Handler) ResolveShortenURL(ctx context.Context, request ResolveShortenURLRequestObject) (ResolveShortenURLResponseObject, error) {
	link, err := h.service.Resolve(ctx, request.ShortCode)
	if err != nil {
		return nil, resolveErrorResponse(err)
	}
	return ResolveShortenURL307Response{Headers: ResolveShortenURL307ResponseHeaders{Location: link.LongURL()}}, nil
}

func createErrorResponse(err error) error {
	switch {
	case errors.Is(err, app.ErrLiveURLLimit):
		return common.NewConflictError("live_url_limit_reached", "user live URL limit reached").WithInternalError(err)
	case errors.Is(err, domain.ErrInvalidLongURL), errors.Is(err, domain.ErrInvalidCode):
		return common.NewInvalidInputError("invalid_shorten_url", "invalid short URL").WithInternalError(err)
	default:
		return common.Error{
			HttpErrorCode: 500,
			PublicError:   "could not create short URL",
			ErrorSlug:     "shorten_url_creation_failed",
			InternalError: err,
		}
	}
}

func resolveErrorResponse(err error) error {
	if errors.Is(err, app.ErrNotFound) || errors.Is(err, domain.ErrInactive) {
		return common.NewNotFoundError("shorten_url_not_found", "short URL not found").WithInternalError(err)
	}
	return common.Error{
		HttpErrorCode: 500,
		PublicError:   "could not resolve short URL",
		ErrorSlug:     "shorten_url_resolution_failed",
		InternalError: err,
	}
}

var _ StrictServerInterface = Handler{}
