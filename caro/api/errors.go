package api

import (
	"errors"

	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	"htqrcode/common"
)

func PublicError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return common.NewInvalidInputError("invalid_caro_input", "invalid room input")
	case errors.Is(err, app.ErrNotFound), errors.Is(err, domain.ErrUnavailable):
		return common.NewNotFoundError("room_unavailable", "room unavailable")
	case errors.Is(err, domain.ErrUnauthorized):
		return common.NewForbiddenError("membership_unauthorized", "membership not authorized")
	case errors.Is(err, app.ErrIdempotencyConflict):
		return common.NewConflictError("idempotency_conflict", "idempotency key was used with different room inputs")
	case errors.Is(err, app.ErrIdempotencyExpired):
		return common.NewExpiredError("idempotency_expired", "the original room has expired; create a new room with a new key")
	case errors.Is(err, domain.ErrWaitingRoomLimit):
		return common.NewConflictError("waiting_room_limit_reached", "an involved IP has reached the limit of 50 waiting or active rooms")
	case errors.Is(err, domain.ErrActiveRoomLimit):
		return common.NewConflictError("active_room_limit_reached", "an involved IP has reached the limit of 50 waiting or active rooms")
	case errors.Is(err, domain.ErrFull):
		return common.NewConflictError("room_full", "room is full")
	case errors.Is(err, app.ErrMembershipConflict):
		return common.NewConflictError("membership_conflict", "connection already belongs to a room")
	case errors.Is(err, app.ErrCollisionExhausted):
		return common.NewConflictError("room_code_exhausted", "could not allocate a room code")
	default:
		return common.Error{HttpErrorCode: 500, PublicError: "could not process room operation", ErrorSlug: "caro_operation_failed", InternalError: err}
	}
}
