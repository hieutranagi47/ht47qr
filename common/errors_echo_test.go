package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

func TestEchoErrorHandler_SerializesCommonError(t *testing.T) {
	e := NewEcho(EchoConfig{HTTPErrorHandler: EchoErrorHandler})
	e.GET("/", func(*echo.Context) error {
		return NewInvalidInputError("invalid-request", "request is invalid").WithDetails([]ErrorDetails{{
			EntityType: "user",
			EntityID:   "123",
			ErrorSlug:  "required",
			Message:    "email is required",
		}, {
			// Required response fields must remain present even when this
			// application error has no value for a field.
		}}).WithInternalError(errors.New("database constraint details"))
	})

	rec := serveRequest(t, e)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.JSONEq(t, `{
		"message":"request is invalid",
		"slug":"invalid-request",
		"details":[{
			"entity_type":"user",
			"entity_id":"123",
			"error_slug":"required",
			"message":"email is required"
		},{
			"entity_type":"",
			"entity_id":"",
			"error_slug":"",
			"message":""
		}]
	}`, rec.Body.String())
	require.NotContains(t, rec.Body.String(), "database constraint details")
}

func TestEchoErrorHandler_UnknownErrorUsesSafeFallback(t *testing.T) {
	e := NewEcho(EchoConfig{})
	e.GET("/", func(*echo.Context) error {
		return errors.New("database password leaked")
	})

	rec := serveRequest(t, e)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.JSONEq(t, `{"message":"Internal Server Error","slug":"internal_server_error","details":[]}`, rec.Body.String())
	require.NotContains(t, rec.Body.String(), "database password leaked")
}

func TestEchoErrorHandler_EchoHTTPErrorUsesSafeStatusResponse(t *testing.T) {
	e := NewEcho(EchoConfig{HTTPErrorHandler: EchoErrorHandler})
	e.GET("/", func(*echo.Context) error {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication implementation detail")
	})

	rec := serveRequest(t, e)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"message":"Unauthorized","slug":"unauthorized","details":[]}`, rec.Body.String())
	require.NotContains(t, rec.Body.String(), "authentication implementation detail")
}

func serveRequest(t *testing.T, e *echo.Echo) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var response HttpErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	return rec
}

func TestEchoErrorHandler_WrappedAndPointerErrors(t *testing.T) {
	structured := NewForbiddenError("access_denied", "access is denied").WithInternalError(errors.New("secret"))
	for _, test := range []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"wrapped value", fmt.Errorf("operation: %w", structured), 403, `{"message":"access is denied","slug":"access_denied","details":[]}`},
		{"pointer", &structured, 403, `{"message":"access is denied","slug":"access_denied","details":[]}`},
		{"wrapped pointer", fmt.Errorf("operation: %w", &structured), 403, `{"message":"access is denied","slug":"access_denied","details":[]}`},
		{"status only", Error{HttpErrorCode: http.StatusBadRequest}, 400, `{"message":"Bad Request","slug":"bad_request","details":[]}`},
		{"wrapped echo", fmt.Errorf("operation: %w", echo.NewHTTPError(405, "secret")), 405, `{"message":"Method Not Allowed","slug":"method_not_allowed","details":[]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := NewEcho(EchoConfig{})
			e.GET("/", func(*echo.Context) error { return test.err })
			rec := serveRequest(t, e)
			require.Equal(t, test.status, rec.Code)
			require.JSONEq(t, test.body, rec.Body.String())
		})
	}
}

func TestEchoErrorHandler_PreservesCommittedResponse(t *testing.T) {
	e := NewEcho(EchoConfig{})
	e.GET("/", func(c *echo.Context) error {
		require.NoError(t, c.String(http.StatusOK, "already sent"))
		return errors.New("late error")
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "already sent", rec.Body.String())
}
