package common

import (
	"errors"
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v5"

	"htqrcode/common/log"
)

func EchoErrorHandler(c *echo.Context, err error) {
	if response, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && response.Committed {
		return
	}

	httpErrorResponse, httpStatus := httpErrorResponseFromErr(err)

	log.FromContext(c.Request().Context()).With("err", err).Error("Handling HTTP error")

	if err := c.JSON(httpStatus, httpErrorResponse); err != nil {
		log.FromContext(c.Request().Context()).With("error", err).Error("Failed to send error response")
	}
}

type HttpErrorResponse struct {
	Message string            `json:"message"`
	Slug    string            `json:"slug"`
	Details []HttpErrorDetail `json:"details"`
}

type HttpErrorDetail struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	ErrorSlug  string `json:"error_slug"`
	Message    string `json:"message"`
}

func httpErrorResponseFromErr(err error) (HttpErrorResponse, int) {
	statusCode := http.StatusInternalServerError
	if code := echo.StatusCode(err); code != 0 {
		statusCode = code
	}

	var commonErr Error
	if !errors.As(err, &commonErr) {
		var commonErrPointer *Error
		if errors.As(err, &commonErrPointer) && commonErrPointer != nil {
			commonErr = *commonErrPointer
		}
	}
	if commonErr.HttpErrorCode != 0 {
		statusCode = commonErr.HttpErrorCode
	}

	publicError := http.StatusText(statusCode)
	errorSlug := strings.ToLower(strings.ReplaceAll(publicError, " ", "_"))
	if commonErr.PublicError != "" {
		publicError = commonErr.PublicError
	}
	if commonErr.ErrorSlug != "" {
		errorSlug = commonErr.ErrorSlug
	}

	httpDetails := make([]HttpErrorDetail, 0, len(commonErr.Details))
	for _, detail := range commonErr.Details {
		httpDetails = append(httpDetails, HttpErrorDetail(detail))
	}

	httpErrorResponse := HttpErrorResponse{
		Slug:    errorSlug,
		Message: publicError,
		Details: httpDetails,
	}

	return httpErrorResponse, statusCode
}

// PublicHTTPError returns the safe error contract shared by REST and gRPC.
func PublicHTTPError(err error) (HttpErrorResponse, int) {
	return httpErrorResponseFromErr(err)
}
